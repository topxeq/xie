package xie

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/topxeq/tk"
)

var versionG string = "0.0.1"

type UndefinedStruct struct {
	int
}

func (o UndefinedStruct) String() string {
	return "undefined"
}

var Undefined UndefinedStruct = UndefinedStruct{0}

var InstrNameSet map[string]int = map[string]int{

	// internal & debug related
	"pass":      101,
	"debug":     102,
	"debugInfo": 103,
	"isUndef":   111,
	"goto":      180,
	"exit":      199,

	// push/peek/pop related
	"push":  220,
	"push$": 221,
	"peek":  222,
	"peek$": 223,
	"pop":   224,
	"pop$":  225,
	"peek*": 226, // from reg
	"pop*":  227, // from reg

	"pushInt":  231,
	"pushInt$": 232,
	"pushInt#": 233,
	"pushInt*": 234,

	// "pushLocal": 290,

	// var related
	"global": 201,
	"var":    203,

	// reg related

	"regInt":  310,
	"regInt#": 312, // from number

	// assign related
	"assign":    401,
	"assign$":   402,
	"assignInt": 410,
	"assignI":   411,

	"assignGlobal": 491,

	// if/else, switch related
	"if":     610,
	"if$":    611,
	"if*":    612,
	"ifNot$": 621,

	// compare related
	"==": 701,

	">i":  710,
	"<i":  720,
	"<i$": 721,
	"<i*": 722,

	// operator related
	"inc":      809,
	"inc$":     810,
	"dec$":     811,
	"dec*":     812,
	"intAdd":   820,
	"intAdd$":  821,
	"intDiv":   831,
	"floatAdd": 840,
	"floatDiv": 848,

	// func related
	"call":     1010,
	"ret":      1020,
	"callFunc": 1050,
	"goFunc":   1060,

	// array related
	"addItem":    1110,
	"addStrItem": 1111,

	// string related
	"backQuote": 1501,
	"quote":     1503,
	"unquote":   1504,
	"trim":      1509,
	"isEmpty":   1510,
	"strAdd":    1520,

	// time related
	"now":           1910,
	"nowStrCompact": 1911,
	"nowStr":        1912,
	"nowStrFormal":  1912,
	"timeSub":       1921,

	// command-line related
	"getParam":       10001,
	"getSwitch":      10002,
	"ifSwitchExists": 10003,

	// print related
	"pln":      10410,
	"plo":      10411,
	"pl":       10420,
	"plv":      10430,
	"plErr":    10440,
	"plErrStr": 10440,

	// convert related
	"convert":  10810,
	"convert$": 10811,

	// err string(TXERROR:) related
	"isErrStr":    10910,
	"getErrStr$":  10922,
	"checkErrStr": 10931,

	// http request/response related
	"writeResp": 20110,

	// web related
	"getWeb": 20210,

	// html related
	"htmlToText": 20310,

	// regex related
	"regReplaceAllStr$": 20411,

	// system related
	"sleep": 20501,
}

type VarRef struct {
	Ref   int // -8 - pop, -7 - peek, -5 - push, -3 - var(string), > 0 normal vars
	Value interface{}
}

type Instr struct {
	Code     int
	ParamLen int
	Params   []VarRef
	// Param1Ref   int
	// Param1Value interface{}
	// Param2Ref   int
	// Param2Value interface{}
}

type Regs struct {
	IntsM [5]int
	CondM bool
}

type FuncContext struct {
	// VarsM          map[int]interface{}
	VarsLocalMapM  map[int]int
	VarsM          *[]interface{}
	ReturnPointerM int
	RegsM          Regs

	// StackM []interface{}
}

type XieVM struct {
	SourceM        []string
	CodeListM      []string
	InstrListM     []Instr
	CodeSourceMapM map[int]int

	LabelsM      map[int]int
	VarIndexMapM map[string]int
	VarNameMapM  map[int]string

	CodePointerM int

	StackM []interface{}

	FuncStackM []FuncContext

	// VarsM map[int]interface{}
	// VarsLocalMapM map[int]int
	// VarsM []interface{}

	// RegsM Regs

	FuncContextM FuncContext

	// CurrentRegsM *Regs
	// // CurrentVarsM *(map[int]interface{})
	// CurrentVarsM *([]interface{})

	CurrentFuncContextM *FuncContext
}

func NewXie(globalsA ...map[string]interface{}) *XieVM {
	vmT := &XieVM{}

	vmT.InitVM(globalsA...)

	return vmT
}

func (p *XieVM) InitVM(globalsA ...map[string]interface{}) {
	p.StackM = make([]interface{}, 0, 10)

	p.FuncStackM = make([]FuncContext, 0, 10)

	p.VarIndexMapM = make(map[string]int, 100)
	p.VarNameMapM = make(map[int]string, 100)

	// p.VarsM = make(map[int]interface{}, 100)
	// p.VarsM = make([]interface{}, 0, 100)
	// p.VarsLocalMapM = make(map[int]string, 100)

	// p.CurrentFuncContextM.RegsM = &(p.RegsM)
	// p.CurrentVarsM = &(p.VarsM)

	// p.FuncContextM = FuncContext{VarsM: make([]interface{}, 0, 10), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: -1}
	p.FuncContextM = FuncContext{VarsM: &([]interface{}{}), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: -1}

	p.CurrentFuncContextM = &(p.FuncContextM)

	p.SetVar("backQuoteG", "`")
	p.SetVar("undefined", Undefined)

	if len(globalsA) > 0 {
		globalsT := globalsA[0]

		for k, v := range globalsT {
			p.SetVar(k, v)
		}
	}

	p.SourceM = make([]string, 0, 100)

	p.CodeListM = make([]string, 0, 100)
	p.InstrListM = make([]Instr, 0, 100)

	p.LabelsM = make(map[int]int, 100)

	p.CodeSourceMapM = make(map[int]int, 100)

}

func (p *XieVM) ParseVar(strA string) VarRef {
	s1T := strings.TrimSpace(strA)

	if strings.HasPrefix(s1T, "`") && strings.HasSuffix(s1T, "`") {
		s1T = s1T[1 : len(s1T)-1]

		return VarRef{-3, s1T} // value(string)
	} else {
		if strings.HasPrefix(s1T, "$") {
			if s1T == "$pop" {
				return VarRef{-8, nil}
			} else if s1T == "$peek" {
				return VarRef{-7, nil}
			} else if s1T == "$push" {
				return VarRef{-5, nil}
			} else {
				vNameT := s1T[1:]

				// if strings.HasPrefix(vNameT, "$") {
				// 	vNameT = vNameT[1:]

				// 	varIndexT, ok := p.VarIndexMapM[vNameT]

				// 	if !ok {
				// 		varIndexT = len(p.VarIndexMapM) + 10000 + 1
				// 		p.VarIndexMapM[vNameT] = varIndexT
				// 		p.VarNameMapM[varIndexT] = vNameT
				// 	}

				// 	return VarRef{varIndexT, nil}
				// }

				varIndexT, ok := p.VarIndexMapM[vNameT]

				if !ok {
					varIndexT = len(p.VarIndexMapM)
					p.VarIndexMapM[vNameT] = varIndexT
					p.VarNameMapM[varIndexT] = vNameT
				}

				return VarRef{varIndexT, nil}
			}
		} else if strings.HasPrefix(s1T, ":") { // labels
			vNameT := s1T[1:]
			varIndexT, ok := p.VarIndexMapM[vNameT]

			if !ok {
				return VarRef{-3, s1T}
			}

			return VarRef{-3, p.LabelsM[varIndexT]}
		} else if strings.HasPrefix(s1T, "#") { // values
			if len(s1T) < 2 {
				return VarRef{-3, s1T}
			}

			// remainsT := s1T[2:]

			typeT := s1T[1]

			if typeT == 'i' {
				c1T, errT := tk.StrToIntQuick(s1T[2:])

				if errT != nil {
					return VarRef{-3, s1T}
				}

				return VarRef{-3, c1T}
			} else if typeT == 'f' {
				c1T, errT := tk.StrToFloat64E(s1T[2:])

				if errT != nil {
					return VarRef{-3, s1T}
				}

				return VarRef{-3, c1T}
			}

			return VarRef{-3, s1T}
		} else {
			return VarRef{-3, s1T} // value(string)
		}
	}
}

// func (p *XieVM) GetVarValue(varA VarRef) interface{} {
// 	idxT := varA.Ref
// 	if idxT == 3 {
// 		return varA.Value
// 	}

// 	if idxT == 8 {
// 		return p.Pop()
// 	}

// 	if idxT == 7 {
// 		return p.Peek()
// 	}

// 	if idxT < 100 {
// 		return fmt.Errorf("invalid var index")
// 	}

// 	return (*(p.CurrentVarsM))[idxT]
// }

func (p *XieVM) GetVarValue(vA VarRef) interface{} {
	idxT := vA.Ref
	if idxT == -3 {
		return vA.Value
	}

	if idxT == -8 {
		return p.Pop()
	}

	if idxT == -7 {
		return p.Peek()
	}

	if idxT == -5 {
		return Undefined
	}

	if idxT < 0 {
		return Undefined
	}

	contextT := p.CurrentFuncContextM

	nv, ok := contextT.VarsLocalMapM[idxT]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[nv]

	// vT, ok := (*(p.CurrentVarsM))[idxT]

	// if !ok {
	// 	return Undefined
	// }

	// return vT
}

func (p *XieVM) GetVarRef(vA VarRef) *interface{} {
	idxT := vA.Ref
	if idxT == -3 {
		return nil
	}

	if idxT == -8 {
		return nil
	}

	if idxT == -7 {
		return nil
	}

	if idxT == -5 {
		return nil
	}

	if idxT < 0 {
		return nil
	}

	// _, ok := p.VarsM[idxT]

	// if !ok {
	// 	return nil
	// }

	contextT := p.CurrentFuncContextM

	return &((*contextT.VarsM)[contextT.VarsLocalMapM[idxT]])
}

func (p *XieVM) GetVarValueGlobal(vA VarRef) interface{} {
	idxT := vA.Ref
	if idxT == -3 {
		return vA.Value
	}

	if idxT == -8 {
		return p.Pop()
	}

	if idxT == -7 {
		return p.Peek()
	}

	if idxT == -5 {
		return Undefined
	}

	if idxT < 0 {
		return Undefined
	}

	contextT := p.FuncContextM

	return (*contextT.VarsM)[contextT.VarsLocalMapM[idxT]]

	// return p.VarsM[idxT]

	// vT, ok := p.VarsM[idxT]

	// if !ok {
	// 	return Undefined
	// }

	// return vT
}

func (p *XieVM) ParseLine(commandA string) ([]string, error) {
	var args []string

	// state: 1 - start, quotes - 2, arg - 3
	state := 1
	current := ""
	quote := "`"
	// escapeNext := false

	command := []rune(commandA)

	for i := 0; i < len(command); i++ {
		c := command[i]

		// if escapeNext {
		// 	current += string(c)
		// 	escapeNext = false
		// 	continue
		// }

		// if c == '\\' {
		// 	current += string(c)
		// 	escapeNext = true
		// 	continue
		// }

		if state == 2 {
			if string(c) != quote {
				current += string(c)
			} else {
				current += string(c) // add it

				args = append(args, current)
				current = ""
				state = 1
			}
			continue
		}

		// tk.Pln(string(c), c, c == '`', '`')
		if c == '"' || c == '\'' || c == '`' {
			state = 2
			quote = string(c)

			current += string(c) // add it

			continue
		}

		if state == 3 {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = 1
			} else {
				current += string(c)
			}
			// Pl("state: %v, current: %v, args: %v", state, current, args)
			continue
		}

		if c != ' ' && c != '\t' {
			state = 3
			current += string(c)
		}
	}

	if state == 2 {
		return []string{}, fmt.Errorf("Unclosed quote in command line: %v", command)
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}

func (p *XieVM) Load(codeA string) string {

	// originSourceLenT := len(p.SourceM)
	originCodeLenT := len(p.CodeListM)

	sourceT := tk.SplitLines(codeA)

	p.SourceM = append(p.SourceM, sourceT...)

	// p.CodeListM = make([]string, 0, len(p.SourceM))
	// p.InstrListM = make([]Instr, 0, len(p.SourceM))

	// p.LabelsM = make(map[int]int, len(p.SourceM))

	// p.CodeSourceMapM = make(map[int]int, len(p.SourceM))

	pointerT := originCodeLenT

	var varCountT int

	for i := 0; i < len(sourceT); i++ {
		v := strings.TrimSpace(sourceT[i])

		if tk.StartsWith(v, "//") || tk.StartsWith(v, "#") {
			continue
		}

		if tk.StartsWith(v, ":") {
			labelT := strings.TrimSpace(v[1:])

			_, ok := p.VarIndexMapM[labelT]

			if !ok {
				varCountT = len(p.VarIndexMapM)

				p.VarIndexMapM[labelT] = varCountT
				p.VarNameMapM[varCountT] = labelT
			}

			p.LabelsM[varCountT] = pointerT

			continue
		}

		iFirstT := i
		if tk.Contains(v, "`") {
			if strings.Count(v, "`")%2 != 0 {
				foundT := false
				var j int
				for j = i + 1; j < len(sourceT); j++ {
					if tk.Contains(sourceT[j], "`") {
						v = tk.JoinLines(sourceT[i : j+1])
						foundT = true
						break
					}
				}

				if !foundT {
					return tk.ErrStrf("parse error: ` not closed(%v)", i)
				}

				i = j
			}
		}

		v = strings.TrimSpace(v)

		if v == "" {
			continue
		}

		p.CodeListM = append(p.CodeListM, v)
		p.CodeSourceMapM[pointerT] = iFirstT
		pointerT++
	}

	for i := originCodeLenT; i < len(p.CodeListM); i++ {
		// listT := strings.SplitN(v, " ", 3)
		v := p.CodeListM[i]
		listT, errT := p.ParseLine(v)
		if errT != nil {
			return p.ErrStrf("failed to parse paramters")
		}

		lenT := len(listT)

		instrNameT := strings.TrimSpace(listT[0])

		codeT, ok := InstrNameSet[instrNameT]

		if !ok {
			return tk.ErrStrf("compile error(line %v/%v %v): unknown instr", i, p.CodeSourceMapM[i]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[i]], 20))
		}

		instrT := Instr{Code: codeT, Params: make([]VarRef, 0, lenT-1)}

		list3T := []VarRef{}

		for j, jv := range listT {
			if j == 0 {
				continue
			}

			list3T = append(list3T, p.ParseVar(jv))
		}

		instrT.Params = append(instrT.Params, list3T...)
		instrT.ParamLen = lenT - 1

		p.InstrListM = append(p.InstrListM, instrT)
	}

	// tk.Plv(p.SourceM)
	// tk.Plv(p.CodeListM)
	// tk.Plv(p.CodeSourceMapM)

	return tk.ToStr(originCodeLenT)
}

func (p *XieVM) PushFunc() {
	// funcContextT := FuncContext{VarsM: make(map[int]interface{}, 10), ReturnPointerM: p.CodePointerM + 1}
	// funcContextT := FuncContext{VarsM: make([]interface{}, 0, 10), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: p.CodePointerM + 1}
	funcContextT := FuncContext{VarsM: &([]interface{}{}), VarsLocalMapM: make(map[int]int, 10), ReturnPointerM: p.CodePointerM + 1}

	p.FuncStackM = append(p.FuncStackM, funcContextT)

	// p.CurrentFuncContextM.RegsM = &(p.FuncStackM[len(p.FuncStackM)-1].RegsM)
	// p.CurrentVarsM = &(p.FuncStackM[len(p.FuncStackM)-1].VarsM)

	p.CurrentFuncContextM = &(p.FuncStackM[len(p.FuncStackM)-1])

}

func (p *XieVM) PopFunc() int {
	funcContextT := p.FuncStackM[len(p.FuncStackM)-1]
	p.FuncStackM = p.FuncStackM[:len(p.FuncStackM)-1]

	if len(p.FuncStackM) < 1 {
		// p.CurrentFuncContextM.RegsM = &(p.RegsM)
		// p.CurrentVarsM = &(p.VarsM)
		p.CurrentFuncContextM = &(p.FuncContextM)
	} else {
		p.CurrentFuncContextM = &(p.FuncStackM[len(p.FuncStackM)-1])
		// p.CurrentFuncContextM.RegsM = &(p.FuncStackM[len(p.FuncStackM)-1].RegsM)
		// p.CurrentVarsM = &(p.FuncStackM[len(p.FuncStackM)-1].VarsM)

	}

	return funcContextT.ReturnPointerM
}

func (p *XieVM) SetVarInt(keyA int, vA interface{}) error {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	if keyA == -5 {
		p.Push(vA)
		return nil
	}

	if keyA < 0 {
		return fmt.Errorf("invalid var index")
	}

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[keyA]

	if !ok {
		localIdxT = len((*contextT.VarsM))

		contextT.VarsLocalMapM[keyA] = localIdxT

		(*contextT.VarsM) = append((*contextT.VarsM), vA)

		// tk.Pln(contextT.VarsM, "***")

		return nil
	}

	// tk.Pln(contextT.VarsLocalMapM, contextT.VarsM, keyA, localIdxT, ok)
	(*contextT.VarsM)[localIdxT] = vA
	// varsT := *(p.CurrentVarsM)

	// idxT, ok := varsT[keyA]

	// varsT[keyA] = vA

	return nil
}

func (p *XieVM) SetVarIntGlobal(keyA int, vA interface{}) error {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	if keyA == -5 {
		p.Push(vA)
		return nil
	}

	if keyA < 0 {
		return fmt.Errorf("invalid var index")
	}

	// contextT := p.FuncContextM

	localIdxT, ok := p.FuncContextM.VarsLocalMapM[keyA]

	if !ok {
		localIdxT = len(*p.FuncContextM.VarsM)

		p.FuncContextM.VarsLocalMapM[keyA] = localIdxT

		*p.FuncContextM.VarsM = append(*p.FuncContextM.VarsM, vA)

		return nil
	}

	(*p.FuncContextM.VarsM)[localIdxT] = vA
	// p.VarsM[keyA] = vA

	return nil
}

func (p *XieVM) SetVar(keyA string, vA interface{}) {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	idxT, ok := p.VarIndexMapM[keyA]
	// tk.Pln(keyA, idxT, ok, p.VarIndexMapM)

	if !ok {
		idxT = len(p.VarIndexMapM)

		p.VarIndexMapM[keyA] = idxT
		p.VarNameMapM[idxT] = keyA

		// tk.Pln(idxT, p.VarIndexMapM, p.VarNameMapM)
	}

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[idxT]

	// tk.Pln(idxT, localIdxT, ok, contextT.VarsLocalMapM)

	if !ok {
		localIdxT = len((*contextT.VarsM))

		contextT.VarsLocalMapM[idxT] = localIdxT

		(*contextT.VarsM) = append((*contextT.VarsM), vA)

		// tk.Pln(idxT, localIdxT, contextT.VarsLocalMapM, contextT.VarsM, "---")

		return
	}

	(*contextT.VarsM)[localIdxT] = vA

	// lenT := len(contextT.VarsM)

	// if lenT < idxT {

	// }

	// varsT := *(p.CurrentVarsM)

	// if len(varsT) < lenT {
	// 	varsT = append(varsT, vA)
	// }

	// ()[lenT] = vA
}

func (p *XieVM) SetVarGlobal(keyA string, vA interface{}) {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	idxT, ok := p.VarIndexMapM[keyA]

	if !ok {
		lenT := len(p.VarIndexMapM) + 1

		p.VarIndexMapM[keyA] = lenT
		p.VarNameMapM[lenT] = keyA

	}

	// contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := p.FuncContextM.VarsLocalMapM[idxT]

	if !ok {
		localIdxT = len(*p.FuncContextM.VarsM)

		p.FuncContextM.VarsLocalMapM[idxT] = localIdxT

		*p.FuncContextM.VarsM = append(*p.FuncContextM.VarsM, vA)

		return
	}

	(*p.FuncContextM.VarsM)[localIdxT] = vA
	// lenT := len(p.VarIndexMapM) + 100

	// p.VarIndexMapM[keyA] = lenT + 1
	// p.VarNameMapM[lenT+1] = keyA

	// p.VarsM[lenT+1] = vA
}

func (p *XieVM) PushVar(vA interface{}) {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	p.Push(vA)
}

func (p *XieVM) GetVarInt(keyA int) interface{} {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[keyA]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[localIdxT]
}

func (p *XieVM) GetVar(keyA string) interface{} {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	idxT, ok := p.VarIndexMapM[keyA]

	if !ok {
		return Undefined

	}

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[idxT]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[localIdxT]

	// lenT := len(p.FuncStackM)

	// if lenT > 0 {
	// 	for i := lenT - 1; i >= 0; i-- {
	// 		varsT := p.FuncStackM[i].VarsM

	// 		vT, ok := varsT[keyA]

	// 		if ok {
	// 			return vT
	// 		}
	// 	}
	// }

	// return p.VarsM[keyA]
}

// get current vars in context
func (p *XieVM) GetVars() []interface{} {
	if p.FuncContextM.VarsM == nil {
		p.InitVM()
	}

	lenT := len(p.FuncStackM)

	if lenT > 0 {
		return *p.FuncStackM[lenT-1].VarsM
	}

	return *p.FuncContextM.VarsM
}

func (p *XieVM) GetRegs() *Regs {
	lenT := len(p.FuncStackM)

	if lenT > 0 {
		return &(p.FuncStackM[lenT-1].RegsM)
	}

	return &(p.FuncContextM.RegsM)
}

func (p *XieVM) Push(vA interface{}) {
	if p.StackM == nil {
		p.InitVM()
	}

	p.StackM = append(p.StackM, vA)
}

// func (p *XieVM) PushLocal(vA interface{}) {
// 	if p.StackM == nil {
// 		p.InitVM()
// 	}

// 	lenT := len(p.FuncStackM)

// 	if lenT > 0 {
// 		p.FuncStackM[lenT-1].StackM = append(p.FuncStackM[lenT-1].StackM, vA)
// 		return
// 	}

// 	p.StackM = append(p.StackM, vA)
// }

func (p *XieVM) Pop() interface{} {
	if p.StackM == nil {
		p.InitVM()

		return Undefined
	}

	lenT := len(p.StackM)

	if lenT < 1 {
		return Undefined
	}

	rs := p.StackM[lenT-1]

	p.StackM = p.StackM[0 : lenT-1]

	return rs
}

// func (p *XieVM) PopLocal() interface{} {
// 	if p.StackM == nil {
// 		p.InitVM()

// 		return Undefined
// 	}

// 	len1T := len(p.FuncStackM)

// 	if len1T > 0 {
// 		lenT := len(p.FuncStackM[len1T-1].StackM)

// 		if lenT < 1 {
// 			return Undefined
// 		}

// 		rs := p.FuncStackM[len1T-1].StackM[lenT-1]

// 		p.FuncStackM[len1T-1].StackM = p.FuncStackM[len1T-1].StackM[0 : lenT-1]

// 		return rs
// 	}

// 	lenT := len(p.StackM)

// 	if lenT < 1 {
// 		return Undefined
// 	}

// 	rs := p.StackM[lenT-1]

// 	p.StackM = p.StackM[0 : lenT-1]

// 	return rs
// }

// func (p *XieVM) PeekLocal() interface{} {
// 	if p.StackM == nil {
// 		p.InitVM()

// 		return Undefined
// 	}

// 	len1T := len(p.FuncStackM)

// 	if len1T > 0 {
// 		lenT := len(p.FuncStackM[len1T-1].StackM)

// 		if lenT < 1 {
// 			return Undefined
// 		}

// 		rs := p.FuncStackM[len1T-1].StackM[lenT-1]

// 		return rs
// 	}

// 	lenT := len(p.StackM)

// 	if lenT < 1 {
// 		return Undefined
// 	}

// 	rs := p.StackM[lenT-1]

// 	return rs
// }

func (p *XieVM) Pops() string {
	if p.StackM == nil {
		p.InitVM()

		return tk.ErrStrf("no value")
	}

	lenT := len(p.StackM)

	if lenT < 1 {
		return tk.ErrStrf("no value")
	}

	rs := p.StackM[lenT-1]

	p.StackM = p.StackM[0 : lenT-1]

	return tk.ToStr(rs)
}

func (p *XieVM) Peek() interface{} {
	if p.StackM == nil {
		p.InitVM()

		return Undefined
	}

	lenT := len(p.StackM)

	if lenT < 1 {
		return Undefined
	}

	return p.StackM[lenT-1]
}

// func (p *XieVM) GetName(nameA string) string {
// 	if tk.StartsWith(nameA, "$") {
// 		return nameA[1:]
// 	} else {
// 		return nameA
// 	}
// }

func (p *XieVM) GetValue(codeA int, vA interface{}) interface{} {
	if codeA == -3 {
		return vA
	}

	if codeA == -8 {
		return p.Pop()
	}

	if codeA == -7 {
		return p.Peek()
	}

	if codeA == -5 {
		return Undefined
	}

	if codeA < 0 {
		return Undefined
	}

	contextT := *(p.CurrentFuncContextM)

	localIdxT, ok := contextT.VarsLocalMapM[codeA]

	if !ok {
		return Undefined
	}

	return (*contextT.VarsM)[localIdxT]
	// return p.VarsM[codeA]

	// vT, ok := p.VarsM[codeA]

	// if !ok {
	// 	return Undefined
	// }

	// return vT
}

// func (p *XieVM) GetValue(nameA string) interface{} {
// 	if tk.StartsWith(nameA, "$") {
// 		nameT := nameA[1:]

// 		if nameT == "pop" {
// 			return p.Pop()
// 		} else if nameT == "peek" {
// 			return p.Peek()
// 		}

// 		return p.GetVars()[nameT]
// 	} else if tk.StartsWith(nameA, `\`) {
// 		return nameA[1:]
// 	} else {
// 		return nameA
// 	}
// }

// func (p *XieVM) Get1Param(strA string) (string, error) {
// 	strT := strings.TrimSpace(strA)

// 	if strT == "" {
// 		return "", tk.Errf("empty")
// 	}

// 	if tk.StartsWith(strT, "`") && tk.EndsWith(strT, "`") {
// 		strT = strT[1 : len(strT)-1]
// 	}

// 	return strT, nil
// }

// func (p *XieVM) Get2Params(strA string) (string, string, error) {
// 	strT := strings.TrimSpace(strA)

// 	if strT == "" {
// 		return "", "", tk.Errf("empty")
// 	}

// 	if tk.StartsWith(strT, "`") {
// 		if tk.EndsWith(strT, "`") {
// 			strT = strT[1 : len(strT)-1]

// 			listT := tk.RegSplitX(strT, "`\\s+`", 2)

// 			if len(listT) < 2 {
// 				return listT[0], "", tk.Errf("not enough parameters")
// 			} else {
// 				return listT[0], listT[1], nil
// 			}
// 		}

// 		if strings.Count(strT, "`") == 2 {
// 			listT := strings.SplitN(strT[1:], "`", 2)

// 			return listT[0], tk.Trim(listT[1]), nil
// 		}

// 	}

// 	listT := strings.SplitN(strT, ` `, 2)

// 	if len(listT) < 2 {
// 		return listT[0], "", tk.Errf("not enough parameters")
// 	}

// 	p2 := strings.TrimSpace(listT[1])
// 	if tk.StartsWith(p2, "`") && tk.EndsWith(p2, "`") {
// 		p2 = p2[1 : len(p2)-1]
// 	}

// 	return listT[0], p2, nil
// }

func (p *XieVM) ErrStrf(formatA string, argsA ...interface{}) string {
	return fmt.Sprintf(fmt.Sprintf("TXERROR:(Line %v: %v) ", p.CodeSourceMapM[p.CodePointerM]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 20))+formatA, argsA...)
}

func (p *XieVM) Debug() {
	tk.Pln(tk.ToJSONX(p, "-indent", "-sort"))
}

func (p *XieVM) RunLine(lineA int) interface{} {
	instrT := p.InstrListM[lineA]

	cmdT := instrT.Code

	switch cmdT {
	case 101: // pass
		return ""
	case 102: // debug
		tk.Pln(tk.ToJSONX(p, "-indent", "-sort"))

		if instrT.ParamLen > 0 {
			if instrT.Params[0].Ref == -3 {
				if instrT.Params[0].Value.(string) == "exit" {
					return "exit"
				}
			}
		}

		return ""
	case 103: // debugInfo
		if instrT.ParamLen < 1 {
			p.Push(tk.ToJSONX(p, "-indent", "-sort"))
			return ""
		}

		nameT := instrT.Params[0].Ref

		if !(nameT == -5 || nameT > 0) {
			return p.ErrStrf("invalid var int")
		}

		p.SetVarInt(nameT, tk.ToJSONX(p, "-indent", "-sort"))

		return ""
	case 111: // isUndef
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		p.Push(v1 == Undefined)

		return ""

	case 180: // goto
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		valueT := p.GetVarValue(instrT.Params[0]).(int)

		return valueT
	case 199: // exit
		if instrT.ParamLen < 1 {
			return "exit"
		}

		valueT := p.GetValue(instrT.Params[0].Ref, instrT.Params[0].Value)

		p.SetVar("outG", valueT)

		return "exit"
	case 201: // global
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		if instrT.ParamLen < 2 {
			p.SetVarInt(instrT.Params[0].Ref, "")
			// p.Curr.VarsM[instrT.Params[0].Ref] = ""
			return ""
		}

		valueT := instrT.Params[1].Value

		if valueT == "bool" {
			p.SetVarInt(instrT.Params[0].Ref, false)
			// p.VarsM[instrT.Params[0].Ref] = false
		} else if valueT == "int" {
			p.SetVarInt(instrT.Params[0].Ref, int(0))
			// p.VarsM[instrT.Params[0].Ref] = int(0)
		} else if valueT == "float" {
			p.SetVarInt(instrT.Params[0].Ref, float64(0.0))
			// p.VarsM[instrT.Params[0].Ref] = float64(0.0)
		} else if valueT == "string" {
			p.SetVarInt(instrT.Params[0].Ref, "")
			// p.VarsM[instrT.Params[0].Ref] = ""
		} else if valueT == "list" {
			p.SetVarInt(instrT.Params[0].Ref, []interface{}{})
			// p.VarsM[instrT.Params[0].Ref] = []interface{}{}
		} else if valueT == "strList" {
			p.SetVarInt(instrT.Params[0].Ref, []string{})
			// p.VarsM[instrT.Params[0].Ref] = []string{}
		} else if valueT == "map" {
			p.SetVarInt(instrT.Params[0].Ref, map[string]interface{}{})
			// p.VarsM[instrT.Params[0].Ref] = map[string]interface{}{}
		} else if valueT == "strMap" {
			p.SetVarInt(instrT.Params[0].Ref, map[string]string{})
			// p.VarsM[instrT.Params[0].Ref] = map[string]string{}
		}

		return ""
	case 203: // var
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		nameT := instrT.Params[0].Ref

		// contextT := p.CurrentFuncContextM

		if instrT.ParamLen < 2 {
			p.SetVarInt(nameT, "")
			// varsT[nameT] = ""
			return ""
		}

		valueT := instrT.Params[1].Value

		if valueT == "bool" {
			p.SetVarInt(nameT, false)
			// varsT[nameT] = false
		} else if valueT == "int" {
			p.SetVarInt(nameT, int(0))
			// varsT[nameT] = int(0)
		} else if valueT == "float" {
			p.SetVarInt(nameT, float64(0.0))
			// varsT[nameT] = float64(0.0)
		} else if valueT == "string" {
			p.SetVarInt(nameT, "")
			// varsT[nameT] = ""
		} else if valueT == "list" {
			p.SetVarInt(nameT, []interface{}{})
			// varsT[nameT] = []interface{}{}
		} else if valueT == "strList" {
			p.SetVarInt(nameT, []string{})
			// varsT[nameT] = []string{}
		} else if valueT == "map" {
			p.SetVarInt(nameT, map[string]interface{}{})
			// varsT[nameT] = map[string]interface{}{}
		} else if valueT == "strMap" {
			p.SetVarInt(nameT, map[string]string{})
			// varsT[nameT] = map[string]string{}
		}

		return ""
	case 220: // push
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		if tk.IsError(v1) {
			return p.ErrStrf("invalid param")
		}

		p.Push(v1)

		return ""
	case 222: // peek
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		p1 := instrT.Params[0].Ref

		if p1 == -5 {
			p.Push(p.Peek())
			return ""
		}

		if p1 < 0 {
			return p.ErrStrf("invalid var name")
		}

		p.SetVarInt(p1, p.Peek())

		// (*(p.CurrentVarsM))[p1] = p.Peek()

		return ""
	case 223: // peek$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		p1 := instrT.Params[0].Ref

		if p1 < 0 {
			return p.ErrStrf("invalid var name")
		}

		p.SetVarInt(p1, p.Peek())
		// (*(p.CurrentVarsM))[p1] = p.Peek()

		return ""
	case 224: // pop
		if instrT.ParamLen < 1 {
			p.Pop()
			return ""
		}

		p1 := instrT.Params[0].Ref

		if p1 < 0 {
			return p.ErrStrf("invalid var name")
		}

		p.SetVarInt(p1, p.Pop())
		// (*(p.CurrentVarsM))[p1] = p.Pop()

		return ""
	case 226: // peek*
		v1 := instrT.Params[0].Value.(int)

		p.CurrentFuncContextM.RegsM.IntsM[v1] = p.Peek().(int)

		return ""
	case 227: // pop*
		v1 := instrT.Params[0].Value.(int)

		p.CurrentFuncContextM.RegsM.IntsM[v1] = p.Pop().(int)

		return ""
	case 231: // pushInt
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		if tk.IsError(v1) {
			return p.ErrStrf("invalid param")
		}

		p.Push(tk.ToInt(v1))

		return ""
	case 232: // pushInt$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}
		// tk.Plv(p.GetVars())

		v1 := p.GetVarValue(instrT.Params[0])

		if tk.IsError(v1) {
			// p.Debug()
			// tk.Plv(instrT)
			return p.ErrStrf("invalid param: %v", v1)
		}

		cT, ok := v1.(int)
		if ok {
			p.Push(cT)
			return ""
		}

		sT, ok := v1.(string)
		if ok {
			c1T, errT := tk.StrToIntQuick(sT)

			if errT != nil {
				return p.ErrStrf("convert value to int failed: %v", errT)
			}

			p.Push(c1T)

			return ""
		}

		return p.ErrStrf("invalid data format")
	case 233: // pushInt#
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := instrT.Params[0].Value.(int)

		// c1T, errT := tk.StrToIntQuick(v1)

		// if errT != nil {
		// 	return p.ErrStrf("convert value to int failed: %v", errT)
		// }

		p.Push(v1)

		return ""
	case 234: // pushInt*
		v1 := instrT.Params[0].Value.(int)

		p.Push(p.CurrentFuncContextM.RegsM.IntsM[v1])

		return ""
	// case 290: // pushLocal
	// 	if instrT.ParamLen < 1 {
	// 		return p.ErrStrf("not enough paramters")
	// 	}

	// 	v1 := p.GetVarValue(instrT.Params[0])

	// 	if tk.IsError(v1) {
	// 		return p.ErrStrf("invalid param")
	// 	}

	// 	p.Push(v1)

	// 	return ""
	case 312: // regInt#  from value
		v1 := instrT.Params[0].Value.(int)
		v2 := instrT.Params[1].Value.(int)

		p.CurrentFuncContextM.RegsM.IntsM[v1] = v2

		return ""
	case 401: // assign
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		nameT := instrT.Params[0].Ref

		if nameT < 0 {
			return p.ErrStrf("invalid var name")
		}

		valueT := p.GetValue(instrT.Params[1].Ref, instrT.Params[1].Value)

		p.SetVarInt(nameT, valueT)

		// (*(p.CurrentVarsM))[nameT] = valueT

		return ""
	case 402: // assign$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		nameT := instrT.Params[0].Ref

		if nameT < 0 {
			return p.ErrStrf("invalid var name")
		}

		p.SetVarInt(nameT, p.Pop())
		// (*(p.CurrentVarsM))[nameT] = p.Pop()

		return ""
	case 410: // assignInt
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		nameT := instrT.Params[0].Ref

		if nameT < 0 {
			return p.ErrStrf("invalid var name")
		}

		valueT := p.GetValue(instrT.Params[1].Ref, instrT.Params[1].Value)

		p.SetVarInt(nameT, tk.ToInt(valueT))

		return ""
	case 411: // assignI
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		p1 := instrT.Params[0].Ref

		if p1 <= 0 {
			return p.ErrStrf("invalid var name")
		}

		v2 := instrT.Params[1].Value.(string)

		c2T, errT := tk.StrToIntQuick(v2)

		if errT != nil {
			return p.ErrStrf("convert value to int failed: %v", errT)
		}

		p.SetVarInt(p1, c2T)

		return ""
	case 491: // assignGlobal
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		nameT := instrT.Params[0].Ref

		if nameT < 0 {
			return p.ErrStrf("invalid var name")
		}

		valueT := p.GetValue(instrT.Params[1].Ref, instrT.Params[1].Value)

		p.SetVarInt(nameT, valueT)

		// p.VarsM[nameT] = valueT

		return ""
	case 492: // assignFromGlobal
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		p1 := instrT.Params[0].Ref

		if p1 <= 0 {
			return p.ErrStrf("invalid var name")
		}

		valueT := p.GetVarValueGlobal(instrT.Params[1])

		p.SetVarInt(p1, valueT)

		return ""
	case 610: // if
		// tk.Plv(instrT)
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		condT := p.GetVarValue(instrT.Params[0]).(bool)

		p2 := p.GetVarValue(instrT.Params[1])

		s2, sok := p2.(string)

		if !sok {
			if condT {
				c2, cok := p2.(int)
				if cok {
					return c2
				}
			}
		} else {
			if condT {
				labelPointerT, ok := p.LabelsM[p.VarIndexMapM[s2]]

				if ok {
					return labelPointerT
				}
			}

		}

		return ""

	case 611: // if$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		condT := p.Pop().(bool)

		if condT {
			return p.GetVarValue(instrT.Params[0]).(int)
		}

		return ""

	case 612: // if*
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		condT := p.CurrentFuncContextM.RegsM.CondM

		if condT {
			return p.GetVarValue(instrT.Params[0]).(int)
		}

		return ""

	case 621: // ifNot$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		condT := p.Pop().(bool)

		if !condT {
			return p.GetVarValue(instrT.Params[0]).(int)
		}

		return ""

	case 701: // ==
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0])

		s2 := p.GetVarValue(instrT.Params[0])

		p.Push(s1 == s2)

		return ""

	case 710: // >i
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(int)

		s2, errT := tk.StrToIntQuick(p.GetVarValue(instrT.Params[1]).(string))

		if errT != nil {
			return p.ErrStrf("failed to convert to int: %v", errT)
		}

		p.Push(s1 > s2)

		return ""

	case 720: // <i
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		var errT error

		p1 := p.GetVarValue(instrT.Params[0])

		c1, ok := p1.(int)

		if !ok {
			s1, ok := p1.(string)

			if ok {
				c1, errT = tk.StrToIntQuick(s1)

				if errT != nil {
					return p.ErrStrf("failed to convert to int: %v", errT)
				}
			} else {
				c1 = tk.ToInt(p1)
			}
		}

		p2 := p.GetVarValue(instrT.Params[1])

		c2, ok := p2.(int)

		if !ok {
			s2, ok := p2.(string)

			if ok {
				c2, errT = tk.StrToIntQuick(s2)

				if errT != nil {
					return p.ErrStrf("failed to convert to int: %v", errT)
				}
			} else {
				c2 = tk.ToInt(p2)
			}
		}

		p.Push(c1 < c2)

		return ""

	case 721: // <i$
		p.Push(p.Pop().(int) > p.Pop().(int))

		return ""

	case 722: // <i*
		regsT := p.CurrentFuncContextM.RegsM
		regsT.CondM = regsT.IntsM[0] < regsT.IntsM[1]

		return ""

	case 809: // inc
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		// varsT := (*(p.CurrentVarsM))

		v1 := p.GetVarValue(instrT.Params[0]).(int)
		// v1 := varsT[p1].(int)

		p.SetVarInt(instrT.Params[0].Ref, v1+1)

		// varsT[p1] = v1 + 1

		return ""

	case 811: // dec$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		// varsT := (*(p.CurrentVarsM))

		p1 := instrT.Params[0].Ref
		// v1 := p.GetVarValue(instrT.Params[0])
		v1 := p.GetVarValue(instrT.Params[0]).(int)

		// if tk.IsError(v1) {
		// 	return p.ErrStrf("invalid param: %v", v1)
		// }

		p.SetVarInt(p1, v1-1)
		// varsT[p1] = v1 - 1

		return ""

	case 812: // dec*
		v1 := instrT.Params[0].Value.(int)

		p.CurrentFuncContextM.RegsM.IntsM[v1]--

		return ""

	case 820: // intAdd
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		p.Push(tk.ToInt(v1) + tk.ToInt(v2))

		return ""

	case 821: // intAdd$
		p.Push(p.Pop().(int) + p.Pop().(int))

		return ""

	case 831: // intDiv
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		p.Push(tk.ToInt(v1) / tk.ToInt(v2))

		return ""

	case 840: // floatAdd
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		p.Push(tk.ToFloat(v1) + tk.ToFloat(v2))

		return ""

	case 848: // floatDiv
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		p.Push(tk.ToFloat(v1) / tk.ToFloat(v2))

		return ""

	case 1010: // call
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		p1 := instrT.Params[0].Value.(int)

		// tk.Pln(tk.ToJSONX(p, "-indent", "-sort"))
		// tk.Pln("p1", p1)
		// tk.Exit()
		p.PushFunc()

		return p1

	case 1020: // ret
		pT := p.PopFunc()

		return pT

	case 1050: // callFunc
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		argCountT := 0

		codeT := ""

		if instrT.ParamLen > 1 {
			argCountT = tk.ToInt(p.GetVarValue(instrT.Params[0]))

			codeT = p.GetVarValue(instrT.Params[1]).(string)
		} else {
			codeT = p.GetVarValue(instrT.Params[0]).(string)
		}

		return p.CallFunc(codeT, argCountT)

	case 1060: // goFunc
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		argCountT := 0

		codeT := ""

		if instrT.ParamLen > 1 {
			argCountT = tk.ToInt(p.GetVarValue(instrT.Params[0]))

			codeT = p.GetVarValue(instrT.Params[1]).(string)
		} else {
			codeT = p.GetVarValue(instrT.Params[0]).(string)
		}

		return p.GoFunc(codeT, argCountT)

	case 1110: // addItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		// varsT := (*(p.CurrentVarsM))

		p1 := instrT.Params[0].Ref
		// v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		varsT := p.GetVars()

		varsT[p1] = append((varsT[p1]).([]interface{}), v2)

		return ""

	case 1111: // addStrItem
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		varsT := p.GetVars()

		p1 := instrT.Params[0].Ref
		// v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		varsT[p1] = append((varsT[p1]).([]string), tk.ToStr(v2))

		return ""

	case 1501: // backQuote
		if instrT.ParamLen > 0 {
			p.SetVarInt(instrT.Params[0].Ref, "`")
		}

		p.Push("`")

		return ""
	case 1503: // quote
		if instrT.ParamLen < 1 {
			rs := strconv.Quote(p.Pop().(string))

			p.Push(rs[1 : len(rs)-1])
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(string)

		rs := strconv.Quote(s1)

		p.Push(rs[1 : len(rs)-1])

		return ""
	case 1504: // unquote
		if instrT.ParamLen < 1 {
			rs, errT := strconv.Unquote(`"` + p.Pop().(string) + `"`)

			if errT != nil {
				p.ErrStrf("failed to unquote: %v", errT)
			}

			p.Push(rs)

			return ""
			// return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(string)

		rs, errT := strconv.Unquote(`"` + s1 + `"`)

		if errT != nil {
			p.ErrStrf("failed to unquote: %v", errT)
		}

		p.Push(rs)

		return ""
	case 1509: // trim
		if instrT.ParamLen < 1 {
			p.Push(strings.TrimSpace(tk.ToStr(p.Pop())))
			return ""
		}

		s1 := p.GetVarValue(instrT.Params[0])

		p.Push(strings.TrimSpace(tk.ToStr(s1)))

		return ""

	case 1510: // isEmpty
		if instrT.ParamLen < 1 {
			p.Push(strings.TrimSpace(tk.ToStr(p.Pop())))
			return ""
		}

		v1 := p.GetVarValue(instrT.Params[0]).(string)

		p.Push(v1 == "")

		return ""

	case 1520: // strAdd
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(string)

		s2 := p.GetVarValue(instrT.Params[1]).(string)

		p.Push(s1 + s2)

		return ""

	case 1910: // now
		if instrT.ParamLen < 1 {
			p.Push(time.Now())

			return ""
		}

		p1 := instrT.Params[0].Ref

		if p1 == -5 {
			p.Push(time.Now())
			return ""
		}

		if p1 < 0 {
			return p.ErrStrf("invalid var name")
		}

		p.SetVarInt(p1, time.Now())

		// (*(p.CurrentVarsM))[p1] = time.Now()

		return ""

	case 1911: // nowStrCompact
		if instrT.ParamLen < 1 {
			p.Push(tk.GetNowTimeString())

			return ""
		}

		p1 := instrT.Params[0].Ref

		if p1 == -5 {
			p.Push(tk.GetNowTimeString())
			return ""
		}

		if p1 < 0 {
			return p.ErrStrf("invalid var name")
		}

		p.SetVarInt(p1, tk.GetNowTimeString())
		// (*(p.CurrentVarsM))[p1] = tk.GetNowTimeString()

		return ""

	case 1912: // nowStr/nowStrFormal
		if instrT.ParamLen < 1 {
			p.Push(tk.GetNowTimeStringFormal())

			return ""
		}

		p1 := instrT.Params[0].Ref

		if p1 == -5 {
			p.Push(tk.GetNowTimeStringFormal())
			return ""
		}

		if p1 < 0 {
			return p.ErrStrf("invalid var name")
		}

		p.SetVarInt(p1, tk.GetNowTimeStringFormal())
		// (*(p.CurrentVarsM))[p1] = tk.GetNowTimeStringFormal()

		return ""
	case 1921: // timeSub
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		sd := int(v1.(time.Time).Sub(v2.(time.Time)))

		p.Push(sd / 1000000)

		return ""

	case 10001: // getParam
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		paramT := tk.GetParameter(v1.([]string), tk.ToInt(v2))

		if tk.IsErrStr(paramT) && instrT.ParamLen > 2 {
			v3 := p.GetVarValue(instrT.Params[2])

			p.Push(v3)

			return ""
		}

		p.Push(paramT)

		return ""

	case 10002: // getSwitch
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		paramT := tk.GetSwitch(v1.([]string), v2.(string))

		p.Push(paramT)

		return ""

	case 10003: // ifSwitchExists
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		paramT := tk.IfSwitchExistsWhole(v1.([]string), v2.(string))

		p.Push(paramT)

		return ""

	case 10410: // pln
		list1T := []interface{}{}

		for _, v := range instrT.Params {
			list1T = append(list1T, p.GetVarValue(v))
		}

		fmt.Println(list1T...)

		return ""
	case 10411: // plo
		if instrT.ParamLen < 1 {
			vT := p.Pop()
			tk.Pl("(%T)%v", vT, vT)
			return ""
		}
		valueT := p.GetVarValue(instrT.Params[0])

		tk.Pl("(%T)%v", valueT, valueT)

		return ""
	case 10420: // pl
		list1T := []interface{}{}

		formatT := ""

		for i, v := range instrT.Params {
			if i == 0 {
				formatT = v.Value.(string)
				continue
			}

			list1T = append(list1T, p.GetVarValue(v))
		}

		fmt.Printf(formatT+"\n", list1T...)

		return ""
	case 10430: // plv
		if instrT.ParamLen < 1 {
			tk.Plv(p.Pop)
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0])

		tk.Plv(s1)

		return ""

	case 10440: // plErr
		if instrT.ParamLen < 1 {
			tk.PlErr(p.Pop().(error))
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(error)

		tk.PlErr(s1)

		return ""

	case 10450: // plErrStr
		if instrT.ParamLen < 1 {
			tk.PlErrString(p.Pop().(string))
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetVarValue(instrT.Params[0]).(string)

		tk.PlErrString(s1)

		return ""

	case 10810: // convert
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		if tk.IsError(v1) {
			return p.ErrStrf("invalid param")
		}

		v2 := p.GetVarValue(instrT.Params[1])

		if tk.IsError(v2) {
			return p.ErrStrf("invalid param")
		}

		s2 := v2.(string)

		if s2 == "bool" {
			p.Push(tk.ToBool(v1))
		} else if s2 == "int" {
			p.Push(tk.ToInt(v1))
		} else if s2 == "float" {
			p.Push(tk.ToFloat(v1))
		} else if v1 == "str" || v1 == "string" {
			p.Push(tk.ToStr(v1))
		} else {
			return p.ErrStrf("unknown type")
		}

		return ""

	case 10811: // convert$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.Pop()

		v2 := p.GetVarValue(instrT.Params[1])

		if tk.IsError(v2) {
			return p.ErrStrf("invalid param")
		}

		s2 := v2.(string)

		if s2 == "b" {
			p.Push(tk.ToBool(s1))
		} else if s2 == "i" {
			p.Push(tk.ToInt(s1))
		} else if s2 == "f" {
			p.Push(tk.ToFloat(s1))
		} else if s2 == "s" {
			p.Push(tk.ToStr(s1))
		} else {
			return p.ErrStrf("unknown type")
		}

		return ""

	case 10910: // isErrStr
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(string)

		if tk.IsErrStr(v1) {
			p.Push(true)
		} else {
			p.Push(false)
		}

		return ""

	case 10922: // getErrStr$
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(string)

		p.Push(tk.GetErrStr(v1))

		return ""

	case 10931: // checkErrStr
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(string)

		if tk.IsErrStr(v1) {
			// tk.Pln(v1)
			return p.ErrStrf(tk.GetErrStr(v1))
			// return "exit"
		}

		return ""

	case 20110: // writeResp
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(http.ResponseWriter)

		v2 := p.GetVarValue(instrT.Params[1]).(string)

		tk.WriteResponse(v1, v2)

		return ""

	case 20210: // getWeb
		if instrT.ParamLen < 2 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0])

		v2 := p.GetVarValue(instrT.Params[1])

		var listT []interface{} = v2.([]interface{})

		// listT = tk.FromJSONWithDefault(tk.ToStr(s2), []interface{}{}).([]interface{})

		// if listT == nil {
		// 	listT = []interface{}{}
		// }

		rs := tk.DownloadWebPageX(tk.ToStr(v1), listT...)

		p.Push(rs)

		return ""

	case 20310: // htmlToText
		var v2 []string
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		if instrT.ParamLen < 2 {
			v2 = []string{}
		} else {
			v2 = p.GetVarValue(instrT.Params[1]).([]string)
		}

		v1 := p.GetVarValue(instrT.Params[0])

		rs := tk.HTMLToText(tk.ToStr(v1), v2...)

		p.Push(rs)

		return ""

	case 20411: // regReplaceAllStr$
		p1 := p.Pops()
		p2 := p.Pops()
		p3 := p.Pops()

		rs := regexp.MustCompile(p2).ReplaceAllString(p3, p1)

		p.Push(rs)

		return ""

	case 20501: // sleep
		if instrT.ParamLen < 1 {
			return p.ErrStrf("not enough paramters")
		}

		v1 := p.GetVarValue(instrT.Params[0]).(float64)

		tk.Sleep(v1)

		return ""

		// end of switch
	}

	return p.ErrStrf("unknown command")
}

func (p *XieVM) CallFunc(codeA string, argCountA int) string {
	vmT := NewXie()

	// argCountT := p.Pop()

	// if argCountT == Undefined {
	// 	return tk.ErrStrf()
	// }

	for i := 0; i < argCountA; i++ {
		vmT.Push(p.Pop())
	}

	lrs := vmT.Load(codeA)

	if tk.IsErrStr(lrs) {
		return lrs
	}

	rs := vmT.Run()

	tk.Plv(rs)

	if !tk.IsErrStr(rs) {
		argCountT := tk.ToInt(rs) // vmT.Pop().(int)

		for i := 0; i < argCountT; i++ {
			p.Push(vmT.Pop())
		}
	}

	return ""
}

func (p *XieVM) GoFunc(codeA string, argCountA int) string {
	vmT := NewXie()

	// argCountT := p.Pop()

	// if argCountT == Undefined {
	// 	return tk.ErrStrf()
	// }

	for i := 0; i < argCountA; i++ {
		vmT.Push(p.Pop())
	}

	lrs := vmT.Load(codeA)

	if tk.IsErrStr(lrs) {
		return lrs
	}

	go vmT.Run()

	return ""
}

// func (p *XieVM) RunLine(lineA int) interface{} {
// 	lineT := p.CodeListM[lineA]

// 	listT := strings.SplitN(lineT, " ", 2)

// 	cmdT := listT[0]

// 	paramsT := ""

// 	if len(listT) > 1 {
// 		paramsT = strings.TrimSpace(listT[1])
// 	}

// 	if cmdT == "pass" {
// 		return ""
// 	} else if cmdT == "global" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			if p1 == "" {
// 				return p.ErrStrf("not enough paramters")
// 			}
// 		}

// 		nameT := p.GetName(p1)

// 		if p2 == "" {
// 			p.VarsM[nameT] = ""
// 			return ""
// 		}

// 		valueT := p.GetValue(p2)

// 		if valueT == "bool" {
// 			p.VarsM[nameT] = false
// 		} else if valueT == "int" {
// 			p.VarsM[nameT] = int(0)
// 		} else if valueT == "float" {
// 			p.VarsM[nameT] = float64(0.0)
// 		} else if valueT == "string" {
// 			p.VarsM[nameT] = ""
// 		} else if valueT == "list" {
// 			p.VarsM[nameT] = []interface{}{}
// 		} else if valueT == "strList" {
// 			p.VarsM[nameT] = []string{}
// 		} else if valueT == "map" {
// 			p.VarsM[nameT] = map[string]interface{}{}
// 		} else if valueT == "strMap" {
// 			p.VarsM[nameT] = map[string]string{}
// 		}

// 		return ""
// 	} else if cmdT == "var" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			if p1 == "" {
// 				return p.ErrStrf("not enough paramters")
// 			}
// 		}

// 		nameT := p.GetName(p1)

// 		varsT := p.GetVars()

// 		if p2 == "" {
// 			varsT[nameT] = ""
// 			return ""
// 		}

// 		valueT := p.GetValue(p2)

// 		if valueT == "bool" {
// 			varsT[nameT] = false
// 		} else if valueT == "int" {
// 			varsT[nameT] = int(0)
// 		} else if valueT == "float" {
// 			varsT[nameT] = float64(0.0)
// 		} else if valueT == "string" {
// 			varsT[nameT] = ""
// 		} else if valueT == "list" {
// 			varsT[nameT] = []interface{}{}
// 		} else if valueT == "strList" {
// 			varsT[nameT] = []string{}
// 		} else if valueT == "map" {
// 			varsT[nameT] = map[string]interface{}{}
// 		} else if valueT == "strMap" {
// 			varsT[nameT] = map[string]string{}
// 		}

// 		return ""
// 	} else if cmdT == "assign" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = valueT

// 		return ""
// 	} else if cmdT == "$assign" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = p.Pop()

// 		return ""
// 	} else if cmdT == "assignBool" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = tk.ToBool(valueT)

// 		return ""
// 	} else if cmdT == "assignInt" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = tk.ToInt(valueT)

// 		return ""
// 	} else if cmdT == "assignFloat" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = tk.ToFloat(valueT)

// 		return ""
// 	} else if cmdT == "assignStr" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		valueT := p.GetValue(p2)

// 		p.GetVars()[nameT] = tk.ToStr(valueT)

// 		return ""
// 	} else if cmdT == "<i" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		p.Push(tk.ToInt(s1) < tk.ToInt(s2))

// 		return ""
// 	} else if cmdT == ">i" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		// tk.Pln(tk.ToInt(s1), tk.ToInt(s2))

// 		p.Push(tk.ToInt(s1) > tk.ToInt(s2))

// 		return ""
// 	} else if cmdT == "if" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		condT := tk.ToBool(s1)

// 		if condT {
// 			labelPointerT, ok := p.LabelsM[tk.ToStr(s2)]

// 			if ok {
// 				return tk.IntToStr(labelPointerT)
// 			}
// 		}

// 		return ""
// 	} else if cmdT == "intAdd" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		p.Push(tk.ToInt(s1) + tk.ToInt(s2))

// 		return ""
// 	} else if cmdT == "intDiv" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		p.Push(tk.ToInt(s1) / tk.ToInt(s2))

// 		return ""
// 	} else if cmdT == "inc" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetName(p1)
// 		v1 := p.GetValue(p1)

// 		p.GetVars()[s1] = tk.ToInt(v1) + 1

// 		return ""
// 	} else if cmdT == "dec" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetName(p1)
// 		v1 := p.GetValue(p1)

// 		p.GetVars()[s1] = tk.ToInt(v1) - 1

// 		return ""
// 	} else if cmdT == "regReplaceAllStr" {
// 		p1 := p.Pops()
// 		p2 := p.Pops()
// 		p3 := p.Pops()

// 		rs := regexp.MustCompile(p2).ReplaceAllString(p3, p1)

// 		p.Push(rs)

// 		return ""
// 	} else if cmdT == "trim" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.Trim(tk.ToStr(p.Pop())))
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		p.Push(tk.Trim(tk.ToStr(s1)))

// 		return ""
// 	} else if cmdT == "plo" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			vT := p.Pop()
// 			tk.Pl("(%T)%v", vT, vT)
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		valueT := p.GetValue(p1)

// 		tk.Pl("(%T)%v", valueT, valueT)

// 		return ""
// 	} else if cmdT == "pl" {
// 		listT, errT := tk.ParseCommandLine(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("failed to parse paramters")
// 		}

// 		list1T := []interface{}{}

// 		formatT := ""

// 		for i, v := range listT {
// 			if i == 0 {
// 				formatT = v
// 				continue
// 			}
// 			list1T = append(list1T, p.GetValue(v))
// 		}

// 		tk.Pl(formatT, list1T...)

// 		return ""
// 	} else if cmdT == "plv" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			tk.Plv(p.Pop())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		valueT := p.GetValue(p1)

// 		tk.Plv(valueT)

// 		return ""
// 	} else if cmdT == "convert" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		if s2 == "bool" {
// 			p.Push(tk.ToBool(s1))
// 		} else if s2 == "int" {
// 			p.Push(tk.ToInt(s1))
// 		} else if s2 == "float" {
// 			p.Push(tk.ToFloat(s1))
// 		} else if s2 == "int" {
// 			p.Push(tk.ToStr(s1))
// 		} else {
// 			return p.ErrStrf("unknown type")
// 		}

// 		return ""
// 	} else if cmdT == "call" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		labelPointerT, ok := p.LabelsM[nameT]

// 		if !ok {
// 			return p.ErrStrf("invalid label")
// 		}

// 		p.PushFunc()

// 		return labelPointerT
// 	} else if cmdT == "ret" {
// 		pT := p.PopFunc()

// 		return pT
// 	} else if cmdT == "pop" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = p.Pop()
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = p.Pop()

// 		return ""
// 	} else if cmdT == "popBool" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = tk.ToBool(p.Pop())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToBool(p.Pop())

// 		return ""
// 	} else if cmdT == "popInt" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = tk.ToInt(p.Pop())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToInt(p.Pop())

// 		return ""
// 	} else if cmdT == "popFloat" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = tk.ToFloat(p.Pop())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToFloat(p.Pop())

// 		return ""
// 	} else if cmdT == "popStr" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["popG"] = p.Pop()
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToStr(p.Pop())

// 		return ""
// 	} else if cmdT == "peek" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = p.Peek()
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = p.Peek()

// 		return ""
// 	} else if cmdT == "$peek" {
// 		// p1, errT := p.Get1Param(paramsT)
// 		// if errT != nil {
// 		// 	p.VarsM["peekG"] = p.Peek()
// 		// 	return ""
// 		// 	// return p.ErrStrf("not enough paramters")
// 		// }

// 		nameT := paramsT

// 		p.GetVars()[nameT] = p.Peek()

// 		return ""
// 	} else if cmdT == "peekBool" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = tk.ToBool(p.Peek())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToBool(p.Peek())

// 		return ""
// 	} else if cmdT == "peekInt" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = tk.ToInt(p.Peek())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToInt(p.Peek())

// 		return ""
// 	} else if cmdT == "peekFloat" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = tk.ToFloat(p.Peek())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToFloat(p.Peek())

// 		return ""
// 	} else if cmdT == "peekStr" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.VarsM["peekG"] = p.Peek()
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		nameT := p.GetName(p1)

// 		p.GetVars()[nameT] = tk.ToStr(p.Peek())

// 		return ""
// 	} else if cmdT == "push" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(p.Pop())
// 			return ""
// 			// return p.ErrStrf("not enough paramters")
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(valueT)

// 		return ""
// 	} else if cmdT == "pushBool" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToBool(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToBool(valueT))

// 		return ""
// 	} else if cmdT == "pushInt" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToInt(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToInt(valueT))

// 		return ""
// 	} else if cmdT == "pushFloat" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToFloat(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToFloat(valueT))

// 		return ""
// 	} else if cmdT == "pushStr" {
// 		p1, errT := p.Get1Param(paramsT)
// 		if errT != nil {
// 			p.Push(tk.ToStr(p.Pop()))
// 			return ""
// 		}

// 		valueT := p.GetValue(p1)

// 		p.Push(tk.ToStr(valueT))

// 		return ""
// 	} else if cmdT == "getParam" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			return p.ErrStrf("not enough paramters")
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		paramT := tk.GetParameter(s1.([]string), tk.ToInt(s2))

// 		p.Push(paramT)

// 		return ""
// 	} else if cmdT == "getNowStr" {
// 		p1, p2, _ := p.Get2Params(paramsT)

// 		var timeStrT string

// 		if p2 == "formal" {
// 			timeStrT = tk.GetNowTimeStringFormal()
// 		} else {
// 			timeStrT = tk.GetNowTimeString()
// 		}

// 		if p1 == "" {
// 			p.Push(timeStrT)
// 		} else {
// 			s1 := p.GetName(p1)

// 			p.GetVars()[s1] = timeStrT
// 		}

// 		return ""
// 	} else if cmdT == "now" {
// 		p1, _ := p.Get1Param(paramsT)

// 		// var timeStrT string

// 		// if p2 == "formal" {
// 		// 	timeStrT = tk.GetNowTimeStringFormal()
// 		// } else {
// 		// 	timeStrT = tk.GetNowTimeString()
// 		// }

// 		if p1 == "" {
// 			p.Push(time.Now())
// 		} else {
// 			s1 := p.GetName(p1)

// 			p.GetVars()[s1] = time.Now()
// 		}

// 		return ""
// 	} else if cmdT == "timeSub" {
// 		p1, p2, _ := p.Get2Params(paramsT)

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		sd := int(s1.(time.Time).Sub(s2.(time.Time)))

// 		p.Push(sd / 1000000)

// 		return ""
// 	} else if cmdT == "getWeb" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			if p1 == "" {
// 				return p.ErrStrf("not enough paramters")
// 			}
// 		}

// 		s1 := p.GetValue(p1)

// 		s2 := p.GetValue(p2)

// 		var listT []interface{} = s2.([]interface{})

// 		// listT = tk.FromJSONWithDefault(tk.ToStr(s2), []interface{}{}).([]interface{})

// 		// if listT == nil {
// 		// 	listT = []interface{}{}
// 		// }

// 		rs := tk.DownloadWebPageX(tk.ToStr(s1), listT...)

// 		p.Push(rs)

// 		return ""
// 	} else if cmdT == "htmlToText" {
// 		p1, p2, errT := p.Get2Params(paramsT)
// 		if errT != nil {
// 			if p1 == "" {
// 				return p.ErrStrf("not enough paramters")
// 			}
// 		}

// 		s1 := p.GetValue(p1)

// 		var s2 []string

// 		if p2 == "" {
// 			s2 = []string{}
// 		} else {
// 			s2 = (p.GetValue(p2)).([]string)
// 		}

// 		rs := tk.HTMLToText(tk.ToStr(s1), s2...)

// 		p.Push(rs)

// 		return ""
// 	} else if cmdT == "getRuntimeInfo" || cmdT == "getDeInfo" {
// 		p.Push(tk.ToJSONX(p, "-indent", "-sort"))

// 		return ""
// 	}

// 	return p.ErrStrf("unknown command")
// }

func (p *XieVM) Run(posA ...int) string {
	p.CodePointerM = 0
	if len(posA) > 0 {
		p.CodePointerM = posA[0]
	}

	for {
		resultT := p.RunLine(p.CodePointerM)

		c1T, ok := resultT.(int)

		if ok {
			p.CodePointerM = c1T
		} else {
			rs, ok := resultT.(string)

			if !ok {
				return p.ErrStrf("invalid return result: (%T)%v", resultT, resultT)
			}

			if tk.IsErrStr(rs) {
				return tk.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStr(rs))
				// tk.Pl("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
				// break
			}

			if rs == "" {
				p.CodePointerM++

				if p.CodePointerM >= len(p.CodeListM) {
					break
				}
			} else if rs == "exit" {
				break
			} else {
				tmpI := tk.StrToInt(rs)

				if tmpI < 0 || tmpI >= len(p.CodeListM) {
					return p.ErrStrf("command index out of range: %v", p.CodePointerM)
				}

				p.CodePointerM = tmpI
			}

		}

	}

	// tk.Pl(tk.ToJSONX(p, "-indent", "-sort"))

	outIndexT, ok := p.VarIndexMapM["outG"]
	if !ok {
		return tk.ErrStrf("no result")
	}

	return tk.ToStr((*p.FuncContextM.VarsM)[p.FuncContextM.VarsLocalMapM[outIndexT]])

}

func RunCode(codeA string, objA interface{}, optsA ...string) interface{} {
	vmT := NewXie()

	if len(optsA) > 0 {
		vmT.SetVar("argsG", optsA)
	}

	if objA != nil {
		vmT.SetVar("inputG", objA)
	}

	lrs := vmT.Load(codeA)

	if tk.IsErrStr(lrs) {
		return tk.ErrStrToErr(lrs)
	}

	// var argsT []string = tk.JSONToStringArray(tk.GetSwitch(optsA, "-args=", "[]"))

	// if argsT != nil {
	// 	vmT.VarsM["argsG"] = argsT
	// } else {
	// 	vmT.VarsM["argsG"] = []string{}
	// }

	rs := vmT.Run()

	return rs
}

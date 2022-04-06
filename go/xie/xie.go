package xie

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/topxeq/tk"
)

var versionG string = "0.0.1"

type XieVM struct {
	SourceM        []string
	CodeListM      []string
	CodeSourceMapM map[int]int
	LabelsM        map[string]int

	CodePointerM int

	StackM []interface{}

	VarsM map[string]interface{}
}

func NewXie() *XieVM {
	vmT := &XieVM{}

	vmT.InitVM()

	return vmT
}

func (p *XieVM) InitVM() {
	p.StackM = make([]interface{}, 0, 10)
	p.VarsM = make(map[string]interface{}, 10)
}

func (p *XieVM) SetVar(keyA string, vA interface{}) {
	if p.VarsM == nil {
		p.InitVM()
	}

	p.VarsM[keyA] = vA
}

func (p *XieVM) GetVar(keyA string) interface{} {
	if p.VarsM == nil {
		p.InitVM()
	}

	return p.VarsM[keyA]
}

func (p *XieVM) Push(vA interface{}) {
	if p.StackM == nil {
		p.InitVM()
	}

	p.StackM = append(p.StackM, vA)
}

func (p *XieVM) Pop() interface{} {
	if p.StackM == nil {
		p.InitVM()

		return nil
	}

	lenT := len(p.StackM)

	if lenT < 1 {
		return nil
	}

	rs := p.StackM[lenT-1]

	p.StackM = p.StackM[0 : lenT-1]

	return rs
}

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

		return nil
	}

	lenT := len(p.StackM)

	if lenT < 1 {
		return nil
	}

	return p.StackM[lenT-1]
}

func (p *XieVM) Load(codeA string) string {

	p.SourceM = tk.SplitLines(codeA)

	p.CodeListM = make([]string, 0, len(p.SourceM))
	p.LabelsM = make(map[string]int, len(p.SourceM))

	p.CodeSourceMapM = make(map[int]int, len(p.SourceM))

	pointerT := 0
	for i := 0; i < len(p.SourceM); i++ {
		v := p.SourceM[i]

		if tk.StartsWith(v, "//") {
			continue
		}

		if tk.StartsWith(v, ":") {
			labelT := tk.Trim(v[1:])

			p.LabelsM[labelT] = pointerT

			continue
		}

		iFirstT := i
		if tk.Contains(v, "||||") {
			if strings.Count(v, "||||")%2 != 0 {
				foundT := false
				var j int
				for j = i + 1; j < len(p.SourceM); j++ {
					if tk.Contains(p.SourceM[j], "||||") {
						v = tk.JoinLines(p.SourceM[i : j+1])
						foundT = true
						break
					}
				}

				if !foundT {
					return tk.ErrStrf("parse error: |||| not closed(%v)", i)
				}

				i = j
			}
		}

		v = tk.Trim(v)

		if v == "" {
			continue
		}

		p.CodeListM = append(p.CodeListM, v)
		p.CodeSourceMapM[pointerT] = iFirstT
		pointerT++
	}

	// tk.Plv(p.SourceM)
	// tk.Plv(p.CodeListM)
	// tk.Plv(p.CodeSourceMapM)

	return ""
}

func (p *XieVM) GetName(nameA string) string {
	if tk.StartsWith(nameA, "$") {
		return nameA[1:]
	} else {
		return nameA
	}
}

func (p *XieVM) GetValue(nameA string) interface{} {
	if tk.StartsWith(nameA, "$") {
		nameT := nameA[1:]

		if nameT == "pop" {
			return p.Pop()
		} else if nameT == "peek" {
			return p.Peek()
		}

		return p.VarsM[nameT]
	} else if tk.StartsWith(nameA, `\`) {
		return nameA[1:]
	} else {
		return nameA
	}
}

func (p *XieVM) Get1Param(strA string) (string, error) {
	strT := tk.Trim(strA)

	if strT == "" {
		return "", tk.Errf("empty")
	}

	if tk.StartsWith(strT, "||||") && tk.EndsWith(strT, "||||") {
		strT = strT[4 : len(strT)-4]
	}

	return strT, nil
}

func (p *XieVM) Get2Params(strA string) (string, string, error) {
	strT := tk.Trim(strA)

	if strT == "" {
		return "", "", tk.Errf("empty")
	}

	if tk.StartsWith(strT, "||||") {
		if tk.EndsWith(strT, "||||") {
			strT = strT[4 : len(strT)-4]

			listT := tk.RegSplitX(strT, `||||\s+||||`, 2)

			if len(listT) < 2 {
				return listT[0], "", tk.Errf("not enough parameters")
			} else {
				return listT[0], listT[1], nil
			}
		}

		if strings.Count(strT, "||||") == 2 {
			listT := strings.SplitN(strT[4:], `||||`, 2)

			return listT[0], tk.Trim(listT[1]), nil
		}

	}

	listT := strings.SplitN(strT, ` `, 2)

	if len(listT) < 2 {
		return listT[0], "", tk.Errf("not enough parameters")
	}

	p2 := tk.Trim(listT[1])
	if tk.StartsWith(p2, "||||") && tk.EndsWith(p2, "||||") {
		p2 = p2[4 : len(p2)-4]
	}

	return listT[0], p2, nil
}

func (p *XieVM) ErrStrf(formatA string, argsA ...interface{}) string {
	return fmt.Sprintf(fmt.Sprintf("TXERROR:(Line %v: %v) ", p.CodeSourceMapM[p.CodePointerM]+1, tk.LimitString(p.SourceM[p.CodeSourceMapM[p.CodePointerM]], 20))+formatA, argsA...)
}

func (p *XieVM) RunLine(lineA int) string {
	lineT := p.CodeListM[lineA]

	listT := strings.SplitN(lineT, " ", 2)

	cmdT := listT[0]

	paramsT := ""

	if len(listT) > 1 {
		paramsT = tk.Trim(listT[1])
	}

	if cmdT == "pass" {
		return ""
	} else if cmdT == "var" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			if p1 == "" {
				return p.ErrStrf("not enough paramters")
			}
		}

		nameT := p.GetName(p1)

		if p2 == "" {
			p.VarsM[nameT] = ""
			return ""
		}

		valueT := p.GetValue(p2)

		if valueT == "bool" {
			p.VarsM[nameT] = false
		} else if valueT == "int" {
			p.VarsM[nameT] = int(0)
		} else if valueT == "float" {
			p.VarsM[nameT] = float64(0.0)
		} else if valueT == "string" {
			p.VarsM[nameT] = ""
		} else if valueT == "list" {
			p.VarsM[nameT] = []interface{}{}
		} else if valueT == "strList" {
			p.VarsM[nameT] = []string{}
		} else if valueT == "map" {
			p.VarsM[nameT] = map[string]interface{}{}
		} else if valueT == "strMap" {
			p.VarsM[nameT] = map[string]string{}
		}

		return ""
	} else if cmdT == "assign" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		valueT := p.GetValue(p2)

		p.VarsM[nameT] = valueT

		return ""
	} else if cmdT == "assignBool" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		valueT := p.GetValue(p2)

		p.VarsM[nameT] = tk.ToBool(valueT)

		return ""
	} else if cmdT == "assignInt" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		valueT := p.GetValue(p2)

		p.VarsM[nameT] = tk.ToInt(valueT)

		return ""
	} else if cmdT == "assignFloat" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		valueT := p.GetValue(p2)

		p.VarsM[nameT] = tk.ToFloat(valueT)

		return ""
	} else if cmdT == "assignStr" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		valueT := p.GetValue(p2)

		p.VarsM[nameT] = tk.ToStr(valueT)

		return ""
	} else if cmdT == "<i" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		p.Push(tk.ToInt(s1) < tk.ToInt(s2))

		return ""
	} else if cmdT == ">i" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		// tk.Pln(tk.ToInt(s1), tk.ToInt(s2))

		p.Push(tk.ToInt(s1) > tk.ToInt(s2))

		return ""
	} else if cmdT == "if" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		condT := tk.ToBool(s1)

		if condT {
			labelPointerT, ok := p.LabelsM[tk.ToStr(s2)]

			if ok {
				return tk.IntToStr(labelPointerT)
			}
		}

		return ""
	} else if cmdT == "exit" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			return "exit"
		}

		v1 := p.GetValue(p1)
		p.SetVar("OutG", v1)

		return "exit"
	} else if cmdT == "strAdd" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		p.Push(tk.ToStr(s1) + tk.ToStr(s2))

		return ""
	} else if cmdT == "intAdd" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		p.Push(tk.ToInt(s1) + tk.ToInt(s2))

		return ""
	} else if cmdT == "inc" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetName(p1)
		v1 := p.GetValue(p1)

		p.VarsM[s1] = tk.ToInt(v1) + 1

		return ""
	} else if cmdT == "dec" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetName(p1)
		v1 := p.GetValue(p1)

		p.VarsM[s1] = tk.ToInt(v1) - 1

		return ""
	} else if cmdT == "regReplaceAllStr" {
		p1 := p.Pops()
		p2 := p.Pops()
		p3 := p.Pops()

		rs := regexp.MustCompile(p2).ReplaceAllString(p3, p1)

		p.Push(rs)

		return ""
	} else if cmdT == "trim" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.Push(tk.Trim(tk.ToStr(p.Pop())))
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		p.Push(tk.Trim(tk.ToStr(s1)))

		return ""
	} else if cmdT == "pln" {
		listT, errT := tk.ParseCommandLine(paramsT)
		if errT != nil {
			// tk.Pln()
			// return ""
			return p.ErrStrf("failed to parse paramters")
		}

		list1T := []interface{}{}

		for _, v := range listT {
			list1T = append(list1T, p.GetValue(v))
		}

		tk.Pln(list1T...)

		return ""
	} else if cmdT == "plo" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			vT := p.Pop()
			tk.Pl("(%T)%v", vT, vT)
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		valueT := p.GetValue(p1)

		tk.Pl("(%T)%v", valueT, valueT)

		return ""
	} else if cmdT == "pl" {
		listT, errT := tk.ParseCommandLine(paramsT)
		if errT != nil {
			return p.ErrStrf("failed to parse paramters")
		}

		list1T := []interface{}{}

		formatT := ""

		for i, v := range listT {
			if i == 0 {
				formatT = v
				continue
			}
			list1T = append(list1T, p.GetValue(v))
		}

		tk.Pl(formatT, list1T...)

		return ""
	} else if cmdT == "plv" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			tk.Plv(p.Pop())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		valueT := p.GetValue(p1)

		tk.Plv(valueT)

		return ""
	} else if cmdT == "convert" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		if s2 == "bool" {
			p.Push(tk.ToBool(s1))
		} else if s2 == "int" {
			p.Push(tk.ToInt(s1))
		} else if s2 == "float" {
			p.Push(tk.ToFloat(s1))
		} else if s2 == "int" {
			p.Push(tk.ToStr(s1))
		} else {
			return p.ErrStrf("unknown type")
		}

		return ""
	} else if cmdT == "pop" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["popG"] = p.Pop()
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = p.Pop()

		return ""
	} else if cmdT == "popBool" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["popG"] = tk.ToBool(p.Pop())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToBool(p.Pop())

		return ""
	} else if cmdT == "popInt" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["popG"] = tk.ToInt(p.Pop())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToInt(p.Pop())

		return ""
	} else if cmdT == "popFloat" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["popG"] = tk.ToFloat(p.Pop())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToFloat(p.Pop())

		return ""
	} else if cmdT == "popStr" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["popG"] = p.Pop()
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToStr(p.Pop())

		return ""
	} else if cmdT == "peek" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["peekG"] = p.Peek()
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = p.Peek()

		return ""
	} else if cmdT == "peekBool" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["peekG"] = tk.ToBool(p.Peek())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToBool(p.Peek())

		return ""
	} else if cmdT == "peekInt" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["peekG"] = tk.ToInt(p.Peek())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToInt(p.Peek())

		return ""
	} else if cmdT == "peekFloat" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["peekG"] = tk.ToFloat(p.Peek())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToFloat(p.Peek())

		return ""
	} else if cmdT == "peekStr" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.VarsM["peekG"] = p.Peek()
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = tk.ToStr(p.Peek())

		return ""
	} else if cmdT == "push" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.Push(p.Pop())
			return ""
			// return p.ErrStrf("not enough paramters")
		}

		valueT := p.GetValue(p1)

		p.Push(valueT)

		return ""
	} else if cmdT == "pushBool" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.Push(tk.ToBool(p.Pop()))
			return ""
		}

		valueT := p.GetValue(p1)

		p.Push(tk.ToBool(valueT))

		return ""
	} else if cmdT == "pushInt" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.Push(tk.ToInt(p.Pop()))
			return ""
		}

		valueT := p.GetValue(p1)

		p.Push(tk.ToInt(valueT))

		return ""
	} else if cmdT == "pushFloat" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.Push(tk.ToFloat(p.Pop()))
			return ""
		}

		valueT := p.GetValue(p1)

		p.Push(tk.ToFloat(valueT))

		return ""
	} else if cmdT == "pushStr" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			p.Push(tk.ToStr(p.Pop()))
			return ""
		}

		valueT := p.GetValue(p1)

		p.Push(tk.ToStr(valueT))

		return ""
	} else if cmdT == "getParam" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		paramT := tk.GetParameter(s1.([]string), tk.ToInt(s2))

		p.Push(paramT)

		return ""
	} else if cmdT == "addItem" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetName(p1)

		s2 := p.GetValue(p2)

		p.VarsM[s1] = append((p.VarsM[s1]).([]interface{}), s2)

		return ""
	} else if cmdT == "addStrItem" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return p.ErrStrf("not enough paramters")
		}

		s1 := p.GetName(p1)

		s2 := p.GetValue(p2)

		p.VarsM[s1] = append((p.VarsM[s1]).([]string), tk.ToStr(s2))

		return ""
	} else if cmdT == "getWeb" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			if p1 == "" {
				return p.ErrStrf("not enough paramters")
			}
		}

		s1 := p.GetValue(p1)

		s2 := p.GetValue(p2)

		var listT []interface{} = s2.([]interface{})

		// listT = tk.FromJSONWithDefault(tk.ToStr(s2), []interface{}{}).([]interface{})

		// if listT == nil {
		// 	listT = []interface{}{}
		// }

		rs := tk.DownloadWebPageX(tk.ToStr(s1), listT...)

		p.Push(rs)

		return ""
	} else if cmdT == "htmlToText" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			if p1 == "" {
				return p.ErrStrf("not enough paramters")
			}
		}

		s1 := p.GetValue(p1)

		var s2 []string

		if p2 == "" {
			s2 = []string{}
		} else {
			s2 = (p.GetValue(p2)).([]string)
		}

		rs := tk.HTMLToText(tk.ToStr(s1), s2...)

		p.Push(rs)

		return ""
	} else if cmdT == "getRuntimeInfo" || cmdT == "getDeInfo" {
		p.Push(tk.ToJSONX(p, "-indent", "-sort"))

		return ""
	} else if cmdT == "debugInfo" {
		tk.Pln(tk.ToJSONX(p, "-indent", "-sort"))

		p1, _ := p.Get1Param(paramsT)
		if p1 == "exit" {
			return "exit"
		}

		return ""
	}

	return p.ErrStrf("unknown command")
}

func (p *XieVM) Run() string {
	p.CodePointerM = 0

	for {
		rs := p.RunLine(p.CodePointerM)

		if tk.IsErrStr(rs) {
			return tk.ErrStrf("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), tk.GetErrStr(rs))
			// tk.Pl("[%v](xie) runtime error: %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
			break
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

	// tk.Pl(tk.ToJSONX(p, "-indent", "-sort"))

	outT, ok := p.VarsM["OutG"]
	if !ok {
		return tk.ErrStrf("no result")
	}

	return tk.ToStr(outT)

}

func RunCode(codeA string, objA interface{}, optsA ...string) string {
	vmT := NewXie()

	vmT.Load(codeA)

	if len(optsA) > 0 {
		vmT.SetVar("argsG", optsA)
	}

	if objA != nil {
		vmT.SetVar("inputG", objA)
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

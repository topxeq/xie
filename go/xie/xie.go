package xie

import (
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

	StackM []string

	RegsM []string

	VarsM map[string]string

	LastNameM  string
	LastValueM string
}

func NewXie() *XieVM {
	vmT := &XieVM{}

	vmT.initVM()

	return vmT
}

func (p *XieVM) initVM() {
	p.StackM = make([]string, 0, 10)
	p.RegsM = make([]string, 0, 10)
	p.VarsM = make(map[string]string, 10)
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

			p.LabelsM[labelT] = i + 1

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

	tk.Plv(p.SourceM)
	tk.Plv(p.CodeListM)
	tk.Plv(p.CodeSourceMapM)

	return ""
}

func (p *XieVM) GetName(nameA string) string {
	if tk.StartsWith(nameA, "$") {
		p.LastNameM = nameA[1:]
	} else {
		p.LastNameM = nameA
	}

	return p.LastNameM
}

func (p *XieVM) GetValue(nameA string) string {
	if tk.StartsWith(nameA, "$") {
		p.LastValueM = p.VarsM[nameA[1:]]
	} else if tk.StartsWith(nameA, `\`) {
		p.LastValueM = nameA[1:]
	} else {
		p.LastValueM = nameA
	}

	return p.LastValueM
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

func (p *XieVM) RunLine(lineA int) string {
	lineT := p.CodeListM[lineA]

	listT := strings.SplitN(lineT, " ", 2)

	cmdT := listT[0]

	paramsT := ""

	if len(listT) > 1 {
		paramsT = tk.Trim(listT[1])
	}

	if cmdT == "pass" {
		p.LastNameM = ""
		p.LastValueM = ""
		return ""
	} else if cmdT == "var" {
		p1, errT := p.Get1Param(paramsT)
		if errT != nil {
			return tk.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		p.VarsM[nameT] = ""

		return ""
	} else if cmdT == "assign" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return tk.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		valueT := p.GetValue(p2)

		p.VarsM[nameT] = valueT

		return ""
	} else if cmdT == "strAdd" {
		p1, p2, errT := p.Get2Params(paramsT)
		if errT != nil {
			return tk.ErrStrf("not enough paramters")
		}

		nameT := p.GetName(p1)

		valueT := p.GetValue(p2)

		p.VarsM[nameT] = p.VarsM[nameT] + valueT

		return ""
	}

	return tk.ErrStrf("unknown command")
}

func (p *XieVM) Run() string {
	p.CodePointerM = 0

	for {
		rs := p.RunLine(p.CodePointerM)

		if tk.IsErrStr(rs) {
			tk.Pl("[%v](xie) runtime error(line %v): %v", tk.GetNowTimeStringFormal(), p.CodeSourceMapM[p.CodePointerM]+1, tk.GetErrStr(rs))
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
			p.CodePointerM = tk.StrToInt(rs)

			if p.CodePointerM < 0 || p.CodePointerM >= len(p.CodeListM) {
				return tk.ErrStrf("command index out of range")
			}
		}
	}

	tk.Pl(tk.ToJSONX(p, "-indent", "-sort"))

	outT, ok := p.VarsM["OutG"]

	if !ok {
		if p.LastValueM != "" {
			return p.LastValueM
		}

		return tk.ErrStrf("no result")
	}

	return tk.ToStr(outT)

}

func getParam(strA string) string {
	strT := tk.Trim(strA)

	if tk.StartsWith(strT, "||||") && tk.EndsWith(strT, "||||") {
		return strT[4 : len(strT)-4]
	}

	return strT
}

func RunLine1(vmA *map[string]interface{}, lineA string, optsA ...string) string {
	vmT := *vmA

	listT := strings.SplitN(lineA, " ", 3)

	cmdLenT := len(listT)

	if cmdLenT < 1 {
		return tk.ErrStrf("empty code line")
	}

	cmdT := tk.Trim(listT[0])

	param1T := ""

	if cmdLenT > 1 {
		param1T = tk.Trim(listT[1])
	}

	param2T := ""

	if cmdLenT > 2 {
		param2T = tk.Trim(listT[2])
	}

	tk.Pl("run line: %v %v %v", cmdT, param1T, param2T)

	if cmdT == "pass" {
		vmT["LastP"] = ""
		vmT["LastV"] = ""
	} else if cmdT == "var" {
		if param1T == "" {
			return tk.ErrStrf("not enough paramters")
		}

		vmT[param1T] = ""

		vmT["LastP"] = param1T
	} else if cmdT == "assign" {
		if cmdLenT < 3 {
			return tk.ErrStrf("not enough paramters")
		}

		if tk.StartsWith(param1T, "$") {
			param1T = param1T[1:]
		}

		if tk.StartsWith(param2T, "$") {
			param2T = vmT[param2T[1:]].(string)
		} else if tk.StartsWith(param2T, "\\") {
			param2T = param2T[1:]
		} else {
			param2T = getParam(param2T)
		}

		vmT[param1T] = param2T

		vmT["LastP"] = param1T
		vmT["LastV"] = param2T
	} else if cmdT == "strAdd" {
		if cmdLenT < 3 {
			return tk.ErrStrf("not enough paramters")
		}

		if tk.StartsWith(param1T, "$") {
			param1T = param1T[1:]
		}

		if tk.StartsWith(param2T, "$") {
			param2T = vmT[param2T[1:]].(string)
		} else if tk.StartsWith(param2T, "\\") {
			param2T = param2T[1:]
		} else {
			param2T = getParam(param2T)
		}

		vmT[param1T] = vmT[param1T].(string) + param2T

		vmT["LastP"] = param1T
		vmT["LastV"] = vmT[param1T]
	}

	return ""
}

// func RunCode1(codeA string, optsA ...string) string {
// 	vmT := make(map[string]interface{})

// 	originalCodeListM := tk.SplitLines(codeA)

// 	codeListM := make([]string, 0, len(originalCodeListM))

// 	codeToOriginMapM := make(map[string]string, len(originalCodeListM))

// 	pointerT := 0
// 	for i := 0; i < len(originalCodeListM); i++ {
// 		v := originalCodeListM[i]

// 		if tk.StartsWith(v, "//") {
// 			continue
// 		}

// 		iFirstT := i
// 		if tk.Contains(v, "||||") {
// 			if strings.Count(v, "||||") != 2 {
// 				foundT := false
// 				var j int
// 				for j = i + 1; j < len(originalCodeListM); j++ {
// 					if tk.Contains(originalCodeListM[j], "||||") {
// 						v = tk.JoinLines(originalCodeListM[i : j+1])
// 						foundT = true
// 						break
// 					}
// 				}

// 				if !foundT {
// 					return tk.ErrStrf("parse error: |||| not closed(%v)", i)
// 				}

// 				i = j
// 			}
// 		}

// 		v = tk.Trim(v)

// 		if v == "" {
// 			continue
// 		}

// 		codeListM = append(codeListM, v)
// 		codeToOriginMapM[tk.IntToStr(pointerT)] = tk.IntToStr(iFirstT)
// 		pointerT++
// 	}

// 	tk.Plv(originalCodeListM)
// 	tk.Plv(codeListM)
// 	tk.Plv(codeToOriginMapM)

// 	// vmT["GlobalsG"] = make(map[string]interface{}, 20)
// 	vmT["OriginalCodeList"] = originalCodeListM
// 	vmT["CodeList"] = codeListM
// 	vmT["CodeToOriginMap"] = codeToOriginMapM
// 	vmT["LinePointer"] = "0"
// 	vmT["LastP"] = ""
// 	vmT["LastV"] = ""

// 	vmT["RegA"] = ""
// 	vmT["RegB"] = ""
// 	vmT["RegC"] = ""

// 	for i, v := range codeListM {
// 		vmT["LinePointer"] = tk.IntToStr(i)
// 		rs := RunLine1(&vmT, v)

// 		if tk.IsErrStr(rs) {
// 			tk.Pl("error: %v", tk.GetErrStr(rs))
// 			break
// 		}

// 		// tk.Plv(vmT["GlobalsG"])
// 	}

// 	tk.Pl(tk.ToJSONX(vmT, "-indent", "-sort"))

// 	outT, ok := vmT["OutG"]

// 	if !ok {
// 		return tk.ErrStrf("no result")
// 	}

// 	return tk.ToStr(outT)
// }

// func Version() string {
// 	return versionG
// }

// func RunCode(codeA string, optsA ...string) string {
// 	vmT := &XieVM{}

// 	vmT.initVM()

// 	vmT.SourceM = tk.SplitLines(codeA)

// 	vmT.CodeListM = make([]string, 0, len(vmT.SourceM))

// 	vmT.CodeSourceMapM = make(map[int]int, len(vmT.SourceM))

// 	pointerT := 0
// 	for i := 0; i < len(vmT.SourceM); i++ {
// 		v := vmT.SourceM[i]

// 		if tk.StartsWith(v, "//") {
// 			continue
// 		}

// 		iFirstT := i
// 		if tk.Contains(v, "||||") {
// 			if strings.Count(v, "||||") != 2 {
// 				foundT := false
// 				var j int
// 				for j = i + 1; j < len(vmT.SourceM); j++ {
// 					if tk.Contains(vmT.SourceM[j], "||||") {
// 						v = tk.JoinLines(vmT.SourceM[i : j+1])
// 						foundT = true
// 						break
// 					}
// 				}

// 				if !foundT {
// 					return tk.ErrStrf("parse error: |||| not closed(%v)", i)
// 				}

// 				i = j
// 			}
// 		}

// 		v = tk.Trim(v)

// 		if v == "" {
// 			continue
// 		}

// 		vmT.CodeListM = append(vmT.CodeListM, v)
// 		vmT.CodeSourceMapM[tk.IntToStr(pointerT)] = tk.IntToStr(iFirstT)
// 		pointerT++
// 	}

// 	tk.Plv(vmT.SourceM)
// 	tk.Plv(vmT.CodeListM)
// 	tk.Plv(vmT.CodeSourceMapM)

// 	// vmT["GlobalsG"] = make(map[string]interface{}, 20)
// 	// vmT.CodePointerM = 0
// 	// vmT.LastP = ""
// 	// vmT.LastV = ""

// 	for i, v := range vmT.CodeListM {
// 		vmT.CodePointerM = i
// 		rs := RunLine1(&vmT, v)

// 		if tk.IsErrStr(rs) {
// 			tk.Pl("error: %v", tk.GetErrStr(rs))
// 			break
// 		}

// 		// tk.Plv(vmT["GlobalsG"])
// 	}

// 	tk.Pl(tk.ToJSONX(vmT, "-indent", "-sort"))

// 	outT, ok := vmT["OutG"]

// 	if !ok {
// 		return tk.ErrStrf("no result")
// 	}

// 	return tk.ToStr(outT)
// }

func RunCode(codeA string, optsA ...string) string {
	vmT := NewXie()

	vmT.Load(codeA)

	rs := vmT.Run()

	return rs
}

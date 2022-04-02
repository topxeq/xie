package xie

import (
	"strings"

	"github.com/topxeq/tk"
)

var versionG string = "0.0.1"

func RunLine(lineA string, optsA ...string) string {
	return ""
}

func RunCode(codeA string, optsA ...string) string {
	vmT := make(map[string]interface{})

	OriginalCodeListT := tk.SplitLines(codeA)

	vmT["OriginalCodeList"] = OriginalCodeListT

	codeListT := make([]string, 0, len(OriginalCodeListT))

	codeToOriginMapT := make(map[int]int, len(OriginalCodeListT))

	pointerT := 0
	for i := 0; i < len(OriginalCodeListT); i++ {
		v := OriginalCodeListT[i]

		if tk.StartsWith(v, "//") {
			continue
		}

		iFirstT := i
		if tk.Contains(v, "||||") {
			if strings.Count(v, "||||") != 2 {
				foundT := false
				var j int
				for j = i + 1; j < len(OriginalCodeListT); j++ {
					if tk.Contains(OriginalCodeListT[j], "||||") {
						v = tk.JoinLines(OriginalCodeListT[i : j+1])
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

		codeListT = append(codeListT, v)
		codeToOriginMapT[pointerT] = iFirstT
		pointerT++
	}

	tk.Plv(OriginalCodeListT)
	tk.Plv(codeListT)
	tk.Plv(codeToOriginMapT)

	outT, ok := vmT["OutG"]

	if !ok {
		return tk.ErrStrf("no result")
	}

	return tk.ToStr(outT)
}

func Version() string {
	return versionG
}

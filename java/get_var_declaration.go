package java

import (
	"TLExtractor/java/types"
	"strings"
)

func GetVarDeclaration(line types.LineInfo) *types.VarInfo {
	if strings.HasPrefix(line.Line, "public") && strings.HasSuffix(line.Line, ";") && line.Nesting == 1 {
		var varInfo types.VarInfo
		data := strings.Split(line.Line, "=")
		trimFunc := func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ';'
		}
		data[0] = strings.TrimFunc(data[0], trimFunc)
		varDec := strings.Split(data[0], " ")
		varDec = varDec[len(varDec)-2:]
		varInfo.Name = varDec[1]
		varInfo.Type = varDec[0]
		if len(data) > 1 {
			varInfo.Value = strings.TrimFunc(data[1], trimFunc)
		}
		return &varInfo
	}
	return nil
}

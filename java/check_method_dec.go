package java

import (
	"TLExtractor/java/types"
	"strings"
)

func CheckMethodDec(line types.LineInfo, name string) bool {
	if strings.HasSuffix(line.Line, "{") && line.Nesting == 2 {
		declaration := strings.Split(line.Line, "(")[0]
		if content := strings.Split(declaration, " "); len(content) > 1 {
			if content[len(content)-1] == name {
				return true
			}
		}
	}
	return false
}

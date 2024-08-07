package java

import (
	"TLExtractor/java/types"
	"strings"
)

func GetParentClass(line types.LineInfo) string {
	if line.Nesting == 1 && strings.HasSuffix(line.Line, "{") {
		if declaration := strings.Split(line.Line, "extends"); len(declaration) > 1 {
			if content := strings.Split(declaration[1], " "); len(content) > 1 {
				return strings.TrimSpace(content[len(content)-2])
			}
		}
	}
	return ""
}

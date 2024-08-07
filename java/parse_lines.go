package java

import (
	"TLExtractor/java/types"
	"strings"
)

func parseLines(page string) []types.LineInfo {
	var nesting int
	var lines []types.LineInfo
	for _, line := range strings.Split(page, "\n") {
		if strings.Contains(line, "{") {
			nesting += 1
		}
		if strings.Contains(line, "}") {
			nesting -= 1
		}
		lines = append(lines, types.LineInfo{Line: strings.TrimSpace(line), Nesting: nesting})
	}
	return lines
}

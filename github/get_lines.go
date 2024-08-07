package github

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme"
	"strings"
)

func getLines(schemeString string) map[string]int {
	lines := make(map[string]int)
	for line, content := range strings.Split(schemeString, "\n") {
		matches := consts.TLSchemeLineRgx.FindAllStringSubmatch(content, -1)
		if len(matches) == 0 {
			continue
		}
		lines[scheme.ReverseConstructor(matches[0][2])] = line + 1
	}
	return lines
}

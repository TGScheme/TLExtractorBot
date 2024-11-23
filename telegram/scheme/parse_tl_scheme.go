package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"regexp"
	"strconv"
	"strings"
)

func ParseTLScheme(rawScheme string) (*types.TLRemoteScheme, error) {
	var generatedScheme types.TLRemoteScheme
	var isMethodDeclaration bool
	compileParams := regexp.MustCompile(`(\w+):(\S+)`)
	compileVersion := regexp.MustCompile(`// LAYER (\d+)`)
	for _, line := range strings.Split(rawScheme, "\n") {
		line = strings.TrimSpace(line)
		if matches := consts.TLSchemeLineRgx.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			tlBase := types.TLBase{
				ID:   ReverseConstructor(matches[0][2]),
				Type: matches[0][5],
			}
			for _, param := range compileParams.FindAllStringSubmatch(matches[0][4], -1) {
				tlBase.Params = append(tlBase.Params, types.Parameter{
					Name: param[1],
					Type: param[2],
				})
			}
			if isMethodDeclaration {
				generatedScheme.Methods = append(generatedScheme.Methods, &types.TLMethod{
					TLBase: tlBase,
					Method: matches[0][1],
				})
			} else {
				generatedScheme.Constructors = append(generatedScheme.Constructors, &types.TLConstructor{
					TLBase:    tlBase,
					Predicate: matches[0][1],
				})
			}
		} else if line == "---functions---" {
			isMethodDeclaration = true
		} else if line == "---types---" {
			isMethodDeclaration = false
		} else if versionMatches := compileVersion.FindAllStringSubmatch(line, -1); len(versionMatches) > 0 {
			layer, err := strconv.Atoi(versionMatches[0][1])
			if err != nil {
				return nil, err
			}
			generatedScheme.Layer = layer
		}
	}
	return &generatedScheme, nil
}

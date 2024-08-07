package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/http"
	"TLExtractor/telegram/scheme/types"
	"regexp"
	"strings"
)

func getScheme() (*types.TLScheme, error) {
	res := http.ExecuteRequest(consts.TDesktopTL)
	if res.Error != nil {
		return nil, res.Error
	}
	var generatedScheme types.TLScheme
	var isMethodDeclaration bool
	compileParams := regexp.MustCompile("(\\w+):(\\S+)")
	for _, line := range strings.Split(string(res.Read()), "\n") {
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
		}
	}
	return &generatedScheme, nil
}

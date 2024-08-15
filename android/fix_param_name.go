package android

import (
	"TLExtractor/consts"
	"TLExtractor/utils"
)

func fixParamName(name string) string {
	newName := utils.SnakeCase(name)
	for rgx, repl := range consts.BrokenNames {
		newName = rgx.ReplaceAllString(newName, repl)
	}
	return newName
}

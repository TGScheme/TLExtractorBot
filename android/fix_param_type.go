package android

import (
	"TLExtractor/consts"
	"TLExtractor/utils"
)

func fixParamType(name string) string {
	newName := utils.SnakeCase(name)
	for rgx, repl := range consts.BrokenTypes {
		newName = rgx.ReplaceAllString(newName, repl)
	}
	return newName
}

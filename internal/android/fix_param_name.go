package android

import (
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/utils"
)

func fixParamName(name string) string {
	newName := utils.SnakeCase(consts.JadxFieldAliasRgx.ReplaceAllString(name, "$1"))
	for rgx, repl := range consts.BrokenNames {
		newName = rgx.ReplaceAllString(newName, repl)
	}
	return newName
}

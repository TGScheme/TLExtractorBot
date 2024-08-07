package resources

import (
	"TLExtractor/consts"
	"github.com/flosch/pongo2/v6"
	"strings"
)

func Format(varName string, args map[string]any) string {
	fromString, err := pongo2.FromString(consts.Templates[varName])
	if err != nil {
		return ""
	}
	fromString.Options.TrimBlocks = true
	execute, err := fromString.Execute(args)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(execute)
}

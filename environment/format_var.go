package environment

import (
	"TLExtractor/assets"
	"strings"

	"github.com/flosch/pongo2/v6"
)

func FormatVar(varName string, args map[string]any) string {
	fromString, err := pongo2.FromString(assets.Templates[varName])
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

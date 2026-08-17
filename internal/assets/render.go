package assets

import (
	"strings"

	"github.com/flosch/pongo2/v6"
)

func Render(name string, args map[string]any) string {
	tpl, err := pongo2.FromString(Templates[name])
	if err != nil {
		return ""
	}
	tpl.Options.TrimBlocks = true
	out, err := tpl.Execute(args)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

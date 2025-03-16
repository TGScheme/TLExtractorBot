package debug_menu

import (
	"TLExtractor/debug_menu/types"
	"TLExtractor/store_api"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/charmbracelet/huh"
)

func newReleaseSelect(typeName string) *types.ReleaseSelect {
	selector := &types.ReleaseSelect{}
	selector.Select = huh.NewSelect[string]().
		Title(fmt.Sprintf("Select %s Telegram Version:", typeName)).
		Value(&selector.Value).
		OptionsFunc(func() []huh.Option[string] {
			releases, err := store_api.GetReleases()
			if err != nil {
				gologging.Fatal(err)
			}
			selector.ReleaseList = releases
			var namedReleases []string
			for _, release := range releases {
				namedReleases = append(namedReleases, selector.NameFormat(release))
			}
			return huh.NewOptions(namedReleases...)
		}, "release")
	return selector
}

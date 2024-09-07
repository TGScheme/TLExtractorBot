package debug_menu

import (
	"TLExtractor/debug_menu/types"
)

func newSelectors(typeName string) (*types.ReleaseSelect, *types.TDeskSelect) {
	stableRelease := newReleaseSelect(typeName)
	stableTDesk := newTDeskSelect(typeName)
	return stableRelease, stableTDesk
}

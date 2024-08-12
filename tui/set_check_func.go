package tui

import (
	"TLExtractor/tui/types"
	"slices"
)

func (miniApp *MiniApp) SetCheckFunc(checker func(types.CheckType) error, filters ...types.CheckType) {
	miniApp.check = func(checkType types.CheckType) error {
		if slices.Contains(filters, checkType) || len(filters) == 0 {
			return checker(checkType)
		}
		return nil
	}
}

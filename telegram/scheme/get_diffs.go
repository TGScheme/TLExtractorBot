package scheme

import "TLExtractor/telegram/scheme/types"

func GetDiffs(old, new *types.TLFullScheme) *types.TLFullDifferences {
	if old == nil || new == nil {
		return nil
	}
	var diff types.TLFullDifferences
	diff.MainApi = getSchemeDiffs(old.MainApi, new.MainApi)
	diff.E2EApi = getSchemeDiffs(old.E2EApi, new.E2EApi)
	if diff.MainApi == nil && diff.E2EApi == nil {
		return nil
	}
	return &diff
}

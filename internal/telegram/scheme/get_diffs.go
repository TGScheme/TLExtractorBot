package scheme

import "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"

func GetDiffs(old, new *types.TLFullScheme) *types.TLFullDifferences {
	if old == nil || new == nil {
		return nil
	}
	var diff types.TLFullDifferences
	diff.MainApi = getSchemeDiffs(old.MainApi, new.MainApi, old.Layer, new.Layer)
	diff.E2EApi = getSchemeDiffs(old.E2EApi, new.E2EApi, old.Layer, new.Layer)
	if diff.MainApi == nil && diff.E2EApi == nil {
		return nil
	}
	return &diff
}

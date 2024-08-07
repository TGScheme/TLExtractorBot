package scheme

import "TLExtractor/telegram/scheme/types"

func GetStats(diffs *types.TLFullDifferences) types.DifferenceStats {
	var stats types.DifferenceStats
	if diffs.MainApi != nil {
		stats.MainApi = getSchemeStats(diffs.MainApi)
	}
	if diffs.E2EApi != nil {
		stats.E2EApi = getSchemeStats(diffs.E2EApi)
	}
	return stats
}

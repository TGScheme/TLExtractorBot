package scheme

import "TLExtractor/telegram/scheme/types"

func getSchemeStats(schemeDifferences *types.TLSchemeDifferences) types.SchemeStats {
	var stats types.SchemeStats
	for _, diff := range schemeDifferences.ConstructorsDifference {
		if diff.IsNew {
			stats.Additions++
		} else {
			stats.Changes++
		}
	}
	for _, diff := range schemeDifferences.MethodsDifference {
		if diff.IsNew {
			stats.Additions++
		} else {
			stats.Changes++
		}
	}
	stats.Total = stats.Additions + stats.Changes
	return stats
}

package scheme

import "TLExtractor/telegram/scheme/types"

func getSchemeStats(schemeDifferences *types.TLSchemeDifferences) types.SchemeStats {
	var stats types.SchemeStats
	for _, diff := range schemeDifferences.ConstructorsDifference {
		if diff.IsNew {
			stats.Constructors.Additions++
		} else {
			stats.Constructors.Changes++
		}
	}
	for _, diff := range schemeDifferences.MethodsDifference {
		if diff.IsNew {
			stats.Methods.Additions++
		} else {
			stats.Constructors.Changes++
		}
	}
	stats.Total = stats.Constructors.Total() + stats.Methods.Total()
	stats.TotalAdditions = stats.Constructors.Additions + stats.Methods.Additions
	stats.TotalChanges = stats.Constructors.Changes + stats.Methods.Changes
	return stats
}

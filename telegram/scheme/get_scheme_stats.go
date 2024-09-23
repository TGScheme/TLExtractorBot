package scheme

import "TLExtractor/telegram/scheme/types"

func getSchemeStats(schemeDifferences *types.TLSchemeDifferences) types.SchemeStats {
	var stats types.SchemeStats
	for _, diff := range schemeDifferences.ConstructorsDifference {
		if diff.IsNew {
			stats.Constructors.Additions++
		} else if diff.IsDeleted {
			stats.Constructors.Deletions++
		} else {
			stats.Constructors.Changes++
		}
	}
	for _, diff := range schemeDifferences.MethodsDifference {
		if diff.IsNew {
			stats.Methods.Additions++
		} else if diff.IsDeleted {
			stats.Methods.Deletions++
		} else {
			stats.Constructors.Changes++
		}
	}
	stats.TotalAdditions = stats.Constructors.Additions + stats.Methods.Additions
	stats.TotalChanges = stats.Constructors.Changes + stats.Methods.Changes
	stats.TotalDeletions = stats.Constructors.Deletions + stats.Methods.Deletions
	stats.Total = stats.TotalAdditions + stats.TotalChanges + stats.TotalDeletions
	return stats
}

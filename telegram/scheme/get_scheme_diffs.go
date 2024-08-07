package scheme

import "TLExtractor/telegram/scheme/types"

func getSchemeDiffs(old, new types.TLScheme) *types.TLSchemeDifferences {
	var diff types.TLSchemeDifferences
	diff.MethodsDifference = getObjsDiffs(old.Methods, new.Methods)
	diff.ConstructorsDifference = getObjsDiffs(old.Constructors, new.Constructors)
	if len(diff.MethodsDifference) == 0 && len(diff.ConstructorsDifference) == 0 {
		return nil
	}
	return &diff
}

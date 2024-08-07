package scheme

import (
	"TLExtractor/telegram/scheme/types"
)

func mergeSchemes(old, new *types.TLScheme) types.TLScheme {
	var newScheme types.TLScheme
	newScheme.Constructors = mergeObjects(old.Constructors, new.Constructors)
	newScheme.Methods = mergeObjects(old.Methods, new.Methods)
	return newScheme
}

package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"slices"
)

func (ctx *context) fixDeprecations(scheme *types.RawTLScheme) *types.RawTLScheme {
	ctx.syncDep.Lock()
	defer ctx.syncDep.Unlock()
	var newScheme types.RawTLScheme
	for _, constructor := range scheme.Constructors {
		if slices.Contains(consts.UnusedTypes, constructor.Package()) {
			continue
		}
		if !slices.Contains(ctx.removedConstructors, ParseConstructor(constructor.ID)) {
			newScheme.Constructors = append(newScheme.Constructors, constructor)
		}
	}
	for _, method := range scheme.Methods {
		if slices.Contains(consts.UnusedTypes, method.Package()) {
			continue
		}
		if !slices.Contains(ctx.removedConstructors, ParseConstructor(method.ID)) {
			newScheme.Methods = append(newScheme.Methods, method)
		}
	}
	return &newScheme
}

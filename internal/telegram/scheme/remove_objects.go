package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func RemoveObjects(scheme *types.TLFullScheme, ids map[string]bool) int {
	if len(ids) == 0 {
		return 0
	}
	removed := 0
	for _, part := range []*types.TLScheme{&scheme.MainApi, &scheme.E2EApi} {
		constructors := part.Constructors[:0]
		for _, constructor := range part.Constructors {
			if ids[ParseConstructor(constructor.Constructor())] {
				removed++
				continue
			}
			constructors = append(constructors, constructor)
		}
		part.Constructors = constructors

		methods := part.Methods[:0]
		for _, method := range part.Methods {
			if ids[ParseConstructor(method.Constructor())] {
				removed++
				continue
			}
			methods = append(methods, method)
		}
		part.Methods = methods
	}
	return removed
}

package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func KeepObjects(scheme, published *types.TLFullScheme, ids map[string]bool) int {
	if published == nil || len(ids) == 0 {
		return 0
	}
	kept := 0
	for _, part := range []struct{ current, published *types.TLScheme }{
		{&scheme.MainApi, &published.MainApi},
		{&scheme.E2EApi, &published.E2EApi},
	} {
		var count int
		part.current.Constructors, count = keepObjects(
			part.current.Constructors, part.published.Constructors, ids,
			func(object *types.TLConstructor) *types.TLConstructor {
				return &types.TLConstructor{TLBase: object.Clone(), Predicate: object.Predicate}
			},
		)
		kept += count
		part.current.Methods, count = keepObjects(
			part.current.Methods, part.published.Methods, ids,
			func(object *types.TLMethod) *types.TLMethod {
				return &types.TLMethod{TLBase: object.Clone(), Method: object.Method}
			},
		)
		kept += count
	}
	return kept
}

func keepObjects[T types.TLInterface](current, published []T, ids map[string]bool, clone func(T) T) ([]T, int) {
	present := make(map[string]bool, len(current))
	names := make(map[string]bool, len(current))
	for _, object := range current {
		present[ParseConstructor(object.Constructor())] = true
		names[object.Package()] = true
	}
	following := make(map[string][]T)
	anchor, kept := "", 0
	for _, object := range published {
		id := ParseConstructor(object.Constructor())
		if present[id] {
			anchor = id
			continue
		}
		if !ids[id] || names[object.Package()] {
			continue
		}
		names[object.Package()] = true
		following[anchor] = append(following[anchor], clone(object))
		kept++
	}
	if kept == 0 {
		return current, 0
	}
	merged := make([]T, 0, len(current)+kept)
	merged = append(merged, following[""]...)
	for _, object := range current {
		merged = append(merged, object)
		merged = append(merged, following[ParseConstructor(object.Constructor())]...)
	}
	return merged, kept
}

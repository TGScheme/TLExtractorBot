package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
	"maps"
	"slices"

	"github.com/Laky-64/gologging"
)

func (ctx *context) computeRemovedConstructors() {
	ctx.removedConstructors = make([]string, 0)
	checkRemovedConstructors := func(old, new []types.ReleasedConstructor) {
		oldSet := make(map[string]int, len(old))
		newSet := make(map[string]int, len(new))
		for _, v := range old {
			oldSet[ParseConstructor(v.ID)] = 1
		}
		for _, v := range new {
			newSet[ParseConstructor(v.ID)] = 1
		}
		for v := range oldSet {
			if _, ok := newSet[v]; !ok {
				ctx.removedConstructors = append(ctx.removedConstructors, v)
			}
		}

		tmp := ctx.removedConstructors[:0]
		for _, v := range ctx.removedConstructors {
			if _, ok := newSet[v]; !ok {
				tmp = append(tmp, v)
			}
		}
		ctx.removedConstructors = tmp
	}
	layers := slices.Collect(maps.Keys(environment.LocalStorage.ReleasedLayers))
	slices.Sort(layers)
	for i := 1; i < len(layers); i++ {
		previousLayer := environment.LocalStorage.ReleasedLayers[layers[i-1]]
		currentLayer := environment.LocalStorage.ReleasedLayers[layers[i]]
		checkRemovedConstructors(previousLayer.Constructors, currentLayer.Constructors)
		checkRemovedConstructors(previousLayer.Methods, currentLayer.Methods)
	}
	ctx.removedComputed = true
}

func (ctx *context) ensureRemovedConstructors() {
	ctx.syncDep.Lock()
	defer ctx.syncDep.Unlock()
	if ctx.removedComputed {
		return
	}
	if _, _, err := ctx.refreshReleasedLayers(); err != nil {
		gologging.Warn("ensureRemovedConstructors: released-layer refresh failed, using cache: ", err)
	}
	ctx.computeRemovedConstructors()
}

func (ctx *context) removedSet() map[string]bool {
	ctx.syncDep.Lock()
	defer ctx.syncDep.Unlock()
	set := make(map[string]bool, len(ctx.removedConstructors))
	for _, v := range ctx.removedConstructors {
		set[v] = true
	}
	return set
}

func (ctx *context) fixDeprecations(scheme *types.RawTLScheme) *types.RawTLScheme {
	ctx.ensureRemovedConstructors()
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

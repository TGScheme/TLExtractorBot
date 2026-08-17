package scheme

import (
	"slices"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func (ctx *Client) ensureRemovedConstructors() {
	ctx.syncDep.Lock()
	defer ctx.syncDep.Unlock()
	if ctx.removedComputed {
		return
	}
	ids, err := ctx.db.ReleasedStore.ListRemovedConstructors()
	if err != nil {
		gologging.Warn("ensureRemovedConstructors: ", err)
		return
	}
	ctx.removedConstructors = make(map[string]bool, len(ids))
	for _, id := range ids {
		ctx.removedConstructors[ParseConstructor(db.FormatID(id))] = true
	}
	ctx.removedComputed = true
}

func (ctx *Client) removedSet() map[string]bool {
	ctx.syncDep.Lock()
	defer ctx.syncDep.Unlock()
	set := make(map[string]bool, len(ctx.removedConstructors))
	for id := range ctx.removedConstructors {
		set[id] = true
	}
	return set
}

func (ctx *Client) fixDeprecations(scheme *types.RawTLScheme) *types.RawTLScheme {
	ctx.ensureRemovedConstructors()
	removed := ctx.removedSet()
	var newScheme types.RawTLScheme
	for _, constructor := range scheme.Constructors {
		if slices.Contains(consts.UnusedTypes, constructor.Package()) || removed[ParseConstructor(constructor.ID)] {
			continue
		}
		newScheme.Constructors = append(newScheme.Constructors, constructor)
	}
	for _, method := range scheme.Methods {
		if slices.Contains(consts.UnusedTypes, method.Package()) || removed[ParseConstructor(method.ID)] {
			continue
		}
		newScheme.Methods = append(newScheme.Methods, method)
	}
	return &newScheme
}

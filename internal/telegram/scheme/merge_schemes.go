package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func (ctx *Client) mergeSchemes(remote *types.TLRemoteScheme, raw *types.TLScheme, rawLayer int, patchOs types.PatchOS, remoteOrder bool) (*types.RawTLScheme, error) {
	var rawScheme types.RawTLScheme
	isSameLayer := remote.Layer == rawLayer
	ctx.ensureRemovedConstructors()
	removed := ctx.removedSet()
	patches, err := loadPatchSet(ctx.db, patchOs)
	if err != nil {
		return nil, err
	}
	rawScheme.Constructors = mergeObjects(remote.Constructors, raw.Constructors, isSameLayer, patches, remoteOrder, removed)
	rawScheme.Methods = mergeObjects(remote.Methods, raw.Methods, isSameLayer, patches, remoteOrder, removed)
	if err = patches.flush(ctx.db); err != nil {
		return nil, err
	}

	canonSrc := make([]types.TLInterface, 0, len(remote.Constructors)+len(remote.Methods))
	for _, c := range remote.Constructors {
		canonSrc = append(canonSrc, c)
	}
	for _, m := range remote.Methods {
		canonSrc = append(canonSrc, m)
	}
	canonIndex := canonicalTypeIndex(canonSrc)
	predIndex := canonicalPredicateIndex(canonSrc)
	canonicalizeScheme(rawScheme.Constructors, canonIndex, predIndex)
	canonicalizeScheme(rawScheme.Methods, canonIndex, predIndex)
	rawScheme.Layer = rawLayer
	rawScheme.IsSync = remote.Layer == rawLayer
	return &rawScheme, nil
}

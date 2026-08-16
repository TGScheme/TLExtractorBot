package scheme

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
)

func mergeSchemes(remote *types.TLRemoteScheme, raw *types.TLScheme, rawLayer int, patchOs types.PatchOS, remoteOrder bool) *types.RawTLScheme {
	var rawScheme types.RawTLScheme
	isSameLayer := remote.Layer == rawLayer
	Client.ensureRemovedConstructors()
	removed := Client.removedSet()
	rawScheme.Constructors = mergeObjects(remote.Constructors, raw.Constructors, isSameLayer, patchOs, remoteOrder, removed)
	rawScheme.Methods = mergeObjects(remote.Methods, raw.Methods, isSameLayer, patchOs, remoteOrder, removed)

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
	environment.LocalStorage.Commit()
	rawScheme.Layer = rawLayer
	rawScheme.IsSync = remote.Layer == rawLayer
	return &rawScheme
}

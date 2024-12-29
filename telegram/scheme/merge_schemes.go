package scheme

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
)

func mergeSchemes(remote *types.TLRemoteScheme, raw *types.TLScheme, rawLayer int, patchOs types.PatchOS, remoteOrder bool) *types.RawTLScheme {
	var rawScheme types.RawTLScheme
	isSameLayer := remote.Layer == rawLayer
	rawScheme.Constructors = mergeObjects(remote.Constructors, raw.Constructors, isSameLayer, patchOs, remoteOrder)
	rawScheme.Methods = mergeObjects(remote.Methods, raw.Methods, isSameLayer, patchOs, remoteOrder)
	environment.LocalStorage.Commit()
	rawScheme.Layer = rawLayer
	rawScheme.IsSync = remote.Layer == rawLayer
	return &rawScheme
}

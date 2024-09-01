package scheme

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
)

func mergeSchemes(remote *types.TLRemoteScheme, raw *types.TLScheme, rawLayer int, patchOs types.PatchOS) *types.RawTLScheme {
	var rawScheme types.RawTLScheme
	isSameLayer := remote.Layer == rawLayer
	rawScheme.Constructors = mergeObjects(remote.Constructors, raw.Constructors, isSameLayer, patchOs)
	rawScheme.Methods = mergeObjects(remote.Methods, raw.Methods, isSameLayer, patchOs)
	environment.LocalStorage.Commit()
	rawScheme.Layer = rawLayer
	rawScheme.IsSync = remote.Layer == rawLayer
	return &rawScheme
}

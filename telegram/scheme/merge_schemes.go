package scheme

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
)

func mergeSchemes(remote *types.TLRemoteScheme, raw *types.TLScheme, rawLayer int) *types.RawTLScheme {
	var rawScheme types.RawTLScheme
	isSameLayer := remote.Layer == rawLayer
	rawScheme.Constructors = mergeObjects(remote.Constructors, raw.Constructors, isSameLayer)
	rawScheme.Methods = mergeObjects(remote.Methods, raw.Methods, isSameLayer)
	environment.LocalStorage.Commit()
	rawScheme.Layer = rawLayer
	rawScheme.IsSync = remote.Layer == rawLayer
	return &rawScheme
}

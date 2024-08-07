package scheme

import "TLExtractor/telegram/scheme/types"

func NewFromRaw(raw *types.RawTLScheme, isE2E bool) *types.TLScheme {
	var scheme types.TLScheme
	scheme.Constructors = filterObjects(raw.Constructors, isE2E)
	scheme.Methods = filterObjects(raw.Methods, isE2E)
	return &scheme
}

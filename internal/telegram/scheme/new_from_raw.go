package scheme

import "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"

func newFromRaw(raw *types.RawTLScheme, isE2E bool) *types.TLScheme {
	var scheme types.TLScheme
	scheme.Constructors = filterObjects(raw.Constructors, isE2E)
	scheme.Methods = filterObjects(raw.Methods, isE2E)
	return &scheme
}

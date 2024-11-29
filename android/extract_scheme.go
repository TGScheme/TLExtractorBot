package android

import (
	"TLExtractor/telegram/scheme"
	schemeTypes "TLExtractor/telegram/scheme/types"
)

func ExtractScheme() (*schemeTypes.TLFullScheme, error) {
	rawScheme, err := extractRawScheme()
	if err != nil {
		return nil, err
	}
	return scheme.MergeSmartUpstream(rawScheme, schemeTypes.AndroidPatch)
}

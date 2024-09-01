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
	return scheme.MergeUpstream(rawScheme, schemeTypes.AndroidPatch, func(isE2E bool) (*schemeTypes.TLRemoteScheme, error) {
		if isE2E {
			return scheme.GetE2EScheme()
		}
		return scheme.GetScheme()
	})
}

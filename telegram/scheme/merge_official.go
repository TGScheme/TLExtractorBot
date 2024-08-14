package scheme

import (
	"TLExtractor/telegram/scheme/types"
)

func MergeOfficial(rawScheme *types.RawTLScheme, isE2E bool) (*types.RawTLScheme, error) {
	var scheme *types.TLRemoteScheme
	if isE2E {
		e2e, err := getE2EScheme()
		if err != nil {
			return nil, err
		}
		scheme = e2e
	} else {
		main, err := getScheme()
		if err != nil {
			return nil, err
		}
		scheme = main
	}
	return mergeSchemes(scheme, newFromRaw(rawScheme, isE2E), rawScheme.Layer), nil
}

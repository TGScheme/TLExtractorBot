package scheme

import (
	"TLExtractor/telegram/scheme/types"
)

func MergeUpstream(rawScheme *types.RawTLScheme, patchOs types.PatchOS, upstream func(isE2E bool) (*types.TLRemoteScheme, error)) (*types.TLFullScheme, error) {
	mergeUpstream := func(rawScheme *types.RawTLScheme, isE2E bool) (*types.RawTLScheme, error) {
		scheme, err := upstream(isE2E)
		if err != nil {
			return nil, err
		}
		return mergeSchemes(scheme, newFromRaw(rawScheme, isE2E), rawScheme.Layer, patchOs), nil
	}
	mainScheme, err := mergeUpstream(rawScheme, false)
	if err != nil {
		return nil, err
	}
	e2eScheme, err := mergeUpstream(rawScheme, true)
	if err != nil {
		return nil, err
	}
	return &types.TLFullScheme{
		MainApi: mainScheme.TLScheme,
		E2EApi:  e2eScheme.TLScheme,
		Layer:   mainScheme.Layer,
		IsSync:  mainScheme.IsSync,
	}, nil
}

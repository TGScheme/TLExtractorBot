package scheme

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
)

func MergeSmartUpstream(rawScheme *types.RawTLScheme, patchOs types.PatchOS) (*types.TLFullScheme, error) {
	return MergeUpstream(rawScheme, patchOs, true, func(isE2E bool) (*types.TLRemoteScheme, error) {
		err := UpdateUpstreamCache("android", nil)
		if err != nil {
			return nil, err
		}
		if isE2E {
			return &types.TLRemoteScheme{
				TLScheme: environment.LocalStorage.UpstreamLayer.E2EApi,
				Layer:    environment.LocalStorage.UpstreamLayer.Layer,
			}, nil
		}
		return &types.TLRemoteScheme{
			TLScheme: environment.LocalStorage.UpstreamLayer.MainApi,
			Layer:    environment.LocalStorage.UpstreamLayer.Layer,
		}, nil
	})
}

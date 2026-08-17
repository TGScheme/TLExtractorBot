package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func (ctx *Client) MergeSmartUpstream(rawScheme *types.RawTLScheme, patchOs types.PatchOS, branch string) (*types.TLFullScheme, error) {
	return ctx.MergeUpstream(rawScheme, patchOs, true, func(isE2E bool) (*types.TLRemoteScheme, error) {
		if err := ctx.UpdateUpstreamCache("android", nil, branch); err != nil {
			return nil, err
		}
		if isE2E {
			return &types.TLRemoteScheme{
				TLScheme: ctx.upstream.E2EApi,
				Layer:    ctx.upstream.Layer,
			}, nil
		}
		return &types.TLRemoteScheme{
			TLScheme: ctx.upstream.MainApi,
			Layer:    ctx.upstream.Layer,
		}, nil
	})
}

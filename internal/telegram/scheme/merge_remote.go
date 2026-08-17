package scheme

import "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"

func (ctx *Client) MergeRemote(remoteScheme *types.TLRemoteScheme, patchOs types.PatchOS, isSync, remoteOrder bool, upstream func(isE2E bool) (*types.TLRemoteScheme, error)) (*types.TLFullScheme, error) {
	var rawScheme types.RawTLScheme
	rawScheme.Layer = remoteScheme.Layer
	rawScheme.Methods = remoteScheme.Methods
	rawScheme.Constructors = remoteScheme.Constructors
	rawScheme.IsSync = isSync
	return ctx.MergeUpstream(&rawScheme, patchOs, remoteOrder, upstream)
}

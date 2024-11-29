package scheme

import "TLExtractor/telegram/scheme/types"

func MergeRemote(remoteScheme *types.TLRemoteScheme, patchOs types.PatchOS, isSync bool, upstream func(isE2E bool) (*types.TLRemoteScheme, error)) (*types.TLFullScheme, error) {
	var rawScheme types.RawTLScheme
	rawScheme.Layer = remoteScheme.Layer
	rawScheme.Methods = remoteScheme.Methods
	rawScheme.Constructors = remoteScheme.Constructors
	rawScheme.IsSync = isSync
	return MergeUpstream(&rawScheme, patchOs, upstream)
}

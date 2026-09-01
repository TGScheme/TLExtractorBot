package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func (ctx *Client) UpdateUpstreamCache(source string, localScheme *types.TLRemoteScheme, branch string) error {
	var remoteScheme *types.TLRemoteScheme
	var patchOs types.PatchOS
	var err error
	if localScheme == nil {
		tempTdLibScheme, errCache := GetTDLibScheme()
		if errCache != nil {
			return errCache
		}
		tempTDeskScheme, errCache := GetScheme(branch)
		if errCache != nil {
			return errCache
		}
		if tempTdLibScheme.Layer > tempTDeskScheme.Layer {
			patchOs = types.TDLibPatch
			localScheme = tempTdLibScheme
			remoteScheme = tempTDeskScheme
		} else {
			patchOs = types.TDesktopPatch
			localScheme = tempTDeskScheme
			remoteScheme = tempTdLibScheme
		}
	} else if source == "tdesktop" {
		remoteScheme, err = GetTDLibScheme()
		patchOs = types.TDesktopPatch
	} else if source == "tdlib" {
		remoteScheme, err = GetScheme(branch)
		patchOs = types.TDLibPatch
	} else {
		return nil
	}
	if err != nil {
		return err
	}
	upstreamScheme, err := ctx.MergeRemote(localScheme, patchOs, false, patchOs == types.TDLibPatch, func(isE2E bool) (*types.TLRemoteScheme, error) {
		if isE2E {
			return GetE2EScheme()
		}
		return remoteScheme, nil
	})
	if err != nil {
		return err
	}
	ctx.upstream = upstreamScheme
	return nil
}

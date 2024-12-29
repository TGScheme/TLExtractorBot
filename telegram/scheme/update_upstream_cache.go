package scheme

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
)

func UpdateUpstreamCache(source string, localScheme *types.TLRemoteScheme) error {
	var remoteScheme *types.TLRemoteScheme
	var patchOs types.PatchOS
	var err error
	if localScheme == nil {
		tempTdLibScheme, errCache := GetTDLibScheme()
		if errCache != nil {
			return err
		}
		tempTDeskScheme, errCache := GetScheme()
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
		remoteScheme, err = GetScheme()
		patchOs = types.TDLibPatch
	} else {
		return nil
	}
	if err != nil {
		return err
	}
	upstreamScheme, err := MergeRemote(localScheme, patchOs, false, patchOs == types.TDLibPatch, func(isE2E bool) (*types.TLRemoteScheme, error) {
		if isE2E {
			return GetE2EScheme()
		}
		return remoteScheme, nil
	})
	if err != nil {
		return err
	}
	environment.LocalStorage.UpstreamLayer = upstreamScheme
	environment.LocalStorage.Commit()
	return nil
}

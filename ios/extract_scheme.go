package ios

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"github.com/Laky-64/goswift/proxy"
)

func ExtractScheme(file *proxy.Context) (*schemeTypes.TLFullScheme, error) {
	rawScheme, err := extractRawScheme(file)
	if err != nil {
		return nil, err
	}
	return scheme.MergeUpstream(rawScheme, schemeTypes.IOSPatch, func(isE2E bool) (*schemeTypes.TLRemoteScheme, error) {
		var remoteScheme schemeTypes.TLRemoteScheme
		if schemePreview := environment.LocalStorage.PreviewLayer; schemePreview != nil {
			if isE2E {
				remoteScheme.TLScheme = schemePreview.E2EApi
			} else {
				remoteScheme.TLScheme = schemePreview.MainApi
			}
			return &remoteScheme, nil
		}
		if isE2E {
			return scheme.GetE2EScheme()
		}
		return scheme.GetScheme()
	})
}

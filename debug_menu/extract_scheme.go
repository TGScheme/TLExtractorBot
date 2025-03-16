package debug_menu

import (
	"TLExtractor/android"
	"TLExtractor/java/jadx"
	"TLExtractor/store_api"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"TLExtractor/tui"
	"fmt"
)

func extractScheme(miniApp *tui.MiniApp, typeName string) (*schemeTypes.TLFullScheme, error) {
	miniApp.SetLoadingMessage(fmt.Sprintf("Downloading %s apk...", typeName))
	info, err := store_api.GetAppInfo()
	if err != nil {
		return nil, err
	}
	err = store_api.DownloadApk(info)
	if err != nil {
		return nil, err
	}
	miniApp.SetLoadingMessage(fmt.Sprintf("Decompiling %s apk...", typeName))
	err = jadx.Decompile(func(percentage int64) {
		miniApp.SetLoadingMessage(fmt.Sprintf("Decompiling %s apk... %d%%", typeName, percentage))
	})
	if err != nil {
		return nil, err
	}
	scheme, err := android.ExtractScheme()
	if err != nil {
		return nil, err
	}
	return scheme, nil
}

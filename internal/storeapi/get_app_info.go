package storeapi

import (
	"encoding/json"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/storeapi/types"

	"github.com/Laky-64/http"
)

func GetAppInfo() (*types.AppInfo, error) {
	var appInfo types.AppInfo
	res, err := http.ExecuteRequest(consts.TDAndroidBetaAPI)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(res.Body, &appInfo); err != nil {
		return nil, err
	}
	appInfo.VersionCode = appInfo.VersionCode / 10
	return &appInfo, nil
}

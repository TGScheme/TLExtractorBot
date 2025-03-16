package store_api

import (
	"TLExtractor/consts"
	"TLExtractor/store_api/types"
	"encoding/json"
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
	return &appInfo, nil
}

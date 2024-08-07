package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/http"
	"encoding/json"
	"fmt"
)

func getAppInfo() (*types.AppInfo, error) {
	var appInfo types.AppInfo
	res := http.ExecuteRequest(
		fmt.Sprintf(
			consts.AppCenterApi,
			consts.Developer,
			consts.AppName,
			consts.Distribution,
		),
	)
	if res.Error != nil {
		return nil, res.Error
	}
	if err := json.Unmarshal(res.Read(), &appInfo); err != nil {
		return nil, err
	}
	return &appInfo, nil
}

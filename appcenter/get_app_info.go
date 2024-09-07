package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"encoding/json"
	"fmt"
	"github.com/Laky-64/http"
)

func GetAppInfo() (*types.AppInfo, error) {
	var appInfo types.AppInfo
	res, err := http.ExecuteRequest(
		fmt.Sprintf(
			consts.AppCenterApi,
			consts.Developer,
			consts.AppName,
			consts.Distribution,
			fmt.Sprintf("releases/%s", consts.AppCenterAndroidRelease),
		),
	)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(res.Body, &appInfo); err != nil {
		return nil, err
	}
	return &appInfo, nil
}

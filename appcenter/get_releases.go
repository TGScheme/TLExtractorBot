package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"encoding/json"
	"fmt"
	"github.com/Laky-64/http"
)

func GetReleases() ([]types.Release, error) {
	var versions []types.Release
	request, err := http.ExecuteRequest(
		fmt.Sprintf(
			consts.AppCenterApi,
			consts.Developer,
			consts.AppName,
			consts.Distribution,
			"public_releases",
		),
	)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(request.Body, &versions)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

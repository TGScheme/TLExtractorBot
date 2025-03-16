package store_api

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/store_api/types"
	"github.com/Laky-64/http"
	"os"
	"path"
)

func DownloadApk(info *types.AppInfo) error {
	res, err := http.ExecuteRequest(info.FileURL)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(environment.EnvFolder, consts.TempBins), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	return os.WriteFile(path.Join(environment.EnvFolder, consts.TempApk), res.Body, os.ModePerm)
}

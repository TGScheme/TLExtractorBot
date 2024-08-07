package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/http"
	"os"
	"path"
)

func downloadApk(info *types.AppInfo) error {
	res := http.ExecuteRequest(
		info.DownloadURL,
	)
	if res.Error != nil {
		return res.Error
	}
	if err := os.MkdirAll(path.Join(consts.BasePath, consts.TempBins), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	return os.WriteFile(path.Join(consts.BasePath, consts.TempApk), res.Read(), os.ModePerm)
}

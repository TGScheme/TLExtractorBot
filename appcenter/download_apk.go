package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"github.com/Laky-64/http"
	"os"
	"path"
)

func downloadApk(info *types.AppInfo) error {
	res, err := http.ExecuteRequest(
		info.DownloadURL,
	)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(consts.EnvFolder, consts.TempBins), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	return os.WriteFile(path.Join(consts.EnvFolder, consts.TempApk), res.Body, os.ModePerm)
}

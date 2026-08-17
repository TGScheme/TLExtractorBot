package storeapi

import (
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/storeapi/types"
	"os"
	"path"
	"time"

	"github.com/Laky-64/http"
)

func DownloadApk(cfg *config.Config, info *types.AppInfo) error {
	res, err := http.ExecuteRequest(fmt.Sprintf("%s&version=%d", info.FileURL, time.Now().Second()))
	if err != nil {
		return err
	}
	if err = os.MkdirAll(path.Join(cfg.WorkDir, consts.TempBins), os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	return os.WriteFile(path.Join(cfg.WorkDir, consts.TempApk), res.Body, os.ModePerm)
}

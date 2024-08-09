package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/logging"
	"slices"
	"time"
)

func Listen(listener func(update types.UpdateInfo) error) {
	for {
		time.Sleep(consts.CheckInterval)
		info, err := getAppInfo()
		if err != nil {
			logging.Error(err)
			continue
		}
		if info.ID > environment.LocalStorage.LastID {
			if err = downloadApk(info); err != nil {
				logging.Error(err)
				continue
			}
			err = listener(
				types.UpdateInfo{
					VersionName: info.VersionName,
					BuildNumber: info.BuildNumber[:4],
				},
			)
			if err != nil {
				logging.Fatal(err)
			}
			if environment.Debug {
				break
			}
			environment.LocalStorage.LastID = info.ID
			environment.LocalStorage.Commit()
		}
	}
}

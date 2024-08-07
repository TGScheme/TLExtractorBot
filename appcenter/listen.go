package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/utils"
	"slices"
	"time"
)

func Listen(listener func(update types.UpdateInfo) error) {
	for {
		time.Sleep(consts.CheckInterval)
		info, err := getAppInfo()
		if err != nil {
			utils.CrashLog(err, false)
			continue
		}
		if info.ID > utils.LocalStorage.LastID {
			if err = downloadApk(info); err != nil {
				utils.CrashLog(err, false)
				continue
			}
			err := listener(
				types.UpdateInfo{
					VersionName: info.VersionName,
					BuildNumber: info.BuildNumber[:4],
				},
			)
			if err != nil {
				utils.CrashLog(err, true)
			}
			if slices.Contains(utils.ShellFlags, "debug") {
				break
			}
			utils.LocalStorage.LastID = info.ID
			if err = utils.LocalStorage.Commit(); err != nil {
				utils.CrashLog(err, true)
			}
		}
	}
}

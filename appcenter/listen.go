package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/environment"
	"github.com/Laky-64/gologging"
	"time"
)

func Listen(listener func(update types.UpdateInfo) error) {
	for {
		time.Sleep(consts.CheckInterval)
		info, err := GetAppInfo()
		if err != nil {
			gologging.Error(err)
			continue
		}
		if info.ID > environment.LocalStorage.LastID && !environment.IsBuilding() || environment.IsPatch() {
			environment.SetBuildingStatus(true)
			if err = DownloadApk(info); err != nil {
				gologging.Error(err)
				continue
			}
			err = listener(
				types.UpdateInfo{
					VersionName: info.VersionName,
					BuildNumber: info.BuildNumber[:4],
				},
			)
			if err != nil {
				panic(err)
			}
			if environment.Debug {
				break
			}
			environment.LocalStorage.LastID = info.ID
			environment.LocalStorage.Commit()
			environment.SetPatchStatus(false)
			environment.SetBuildingStatus(false)
		}
	}
}

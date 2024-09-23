package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/environment"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"strconv"
	"time"
)

func Listen(listener func(update types.UpdateInfo) error) {
	waitTime := time.Hour / consts.MaxGithubRequests * consts.NumSources
	for {
		time.Sleep(waitTime)
		info, err := GetAppInfo()
		if err != nil {
			gologging.Error(err)
			continue
		}
		if !environment.IsBuilding() {
			request, err := http.ExecuteRequest(fmt.Sprintf(consts.TDesktopSources+"/core/version.h", consts.TDesktopBranch))
			if err != nil {
				gologging.Error(err)
				continue
			}
			body := string(request.Body)
			versionCode, _ := strconv.Atoi(consts.TDeskVersionRgx.FindAllStringSubmatch(body, -1)[0][1])
			versionName := consts.TDeskVersionNameRgx.FindAllStringSubmatch(body, -1)[0][1]
			if versionCode > environment.LocalStorage.LastTDeskID {
				environment.SetBuildingStatus(true)
				err = listener(
					types.UpdateInfo{
						VersionName: versionName,
						BuildNumber: strconv.Itoa(versionCode),
						Source:      "tdesktop",
					},
				)
				if err != nil {
					panic(err)
				}
				if environment.Debug {
					break
				}
				environment.LocalStorage.LastTDeskID = versionCode
				environment.SetBuildingStatus(false)
			}
			if info.ID > environment.LocalStorage.LastID || environment.IsPatch() {
				environment.SetBuildingStatus(true)
				if err = DownloadApk(info); err != nil {
					gologging.Error(err)
					continue
				}
				err = listener(
					types.UpdateInfo{
						VersionName: info.VersionName,
						BuildNumber: info.BuildNumber[:4],
						Source:      "android",
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
}

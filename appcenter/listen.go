package appcenter

import (
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/utils"
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
			tDeskBody := request.String()
			tDeskVersionCode, _ := strconv.Atoi(consts.TDeskVersionRgx.FindAllStringSubmatch(tDeskBody, -1)[0][1])
			tDeskVersionName := consts.TDeskVersionNameRgx.FindAllStringSubmatch(tDeskBody, -1)[0][1]
			if tDeskVersionCode > environment.LocalStorage.LastTDeskID {
				environment.SetBuildingStatus(true)
				err = listener(
					types.UpdateInfo{
						VersionName: tDeskVersionName,
						BuildNumber: strconv.Itoa(tDeskVersionCode),
						Source:      "tdesktop",
					},
				)
				if err != nil {
					gologging.Fatal(err)
				}
				if environment.Debug {
					break
				}
				environment.LocalStorage.LastTDeskID = tDeskVersionCode
				environment.LocalStorage.Commit()
				environment.SetBuildingStatus(false)
			}
			request, err = http.ExecuteRequest(consts.TDLibSources + "/CMakeLists.txt")
			if err != nil {
				gologging.Error(err)
				continue
			}
			tdLibBody := request.String()
			tdLibVersionName := consts.TDLibVersionRgx.FindAllStringSubmatch(tdLibBody, -1)[0][1]
			tdLibVersionCode := utils.VersionToCode(tdLibVersionName)
			if tdLibVersionCode > environment.LocalStorage.LastTDLibID {
				environment.SetBuildingStatus(true)
				err = listener(
					types.UpdateInfo{
						VersionName: tdLibVersionName,
						BuildNumber: strconv.Itoa(tdLibVersionCode),
						Source:      "tdlib",
					},
				)
				if err != nil {
					gologging.Fatal(err)
				}
				if environment.Debug {
					break
				}
				environment.LocalStorage.LastTDLibID = tdLibVersionCode
				environment.LocalStorage.Commit()
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
					gologging.Fatal(err)
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

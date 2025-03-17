package store_api

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/store_api/types"
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
						BuildNumber: uint32(tDeskVersionCode),
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
						BuildNumber: uint32(tdLibVersionCode),
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
			if info.VersionCode > environment.LocalStorage.LastVersionCode || environment.IsPatch() {
				environment.SetBuildingStatus(true)
				if err = DownloadApk(info); err != nil {
					gologging.Error(err)
					continue
				}
				err = listener(
					types.UpdateInfo{
						VersionName: info.Version,
						BuildNumber: info.VersionCode,
						Source:      "android",
					},
				)
				if err != nil {
					gologging.Fatal(err)
				}
				if environment.Debug {
					break
				}
				environment.LocalStorage.LastVersionCode = info.VersionCode
				environment.LocalStorage.Commit()
				environment.SetPatchStatus(false)
				environment.SetBuildingStatus(false)
			}
		}
	}
}

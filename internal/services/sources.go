package services

import (
	"fmt"
	"runtime/debug"

	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/storeapi"
	storeTypes "github.com/TGScheme/TLExtractorBot/internal/storeapi/types"
	"github.com/TGScheme/TLExtractorBot/internal/utils"
)

func (s *Service) pollSources() {
	if s.building.Load() {
		return
	}
	settings, err := s.db.SettingsStore.GetSettings()
	if err != nil {
		gologging.Error(err)
		return
	}
	hasPreview, err := s.hasPreview()
	if err != nil {
		gologging.Error(err)
		return
	}

	if hasPreview {
		if version, name, errFetch := s.tdesktopVersion(settings.TdesktopBranch); errFetch != nil {
			gologging.Error(errFetch)
		} else if int64(version) > settings.LastTdeskID {
			s.dispatch(storeTypes.UpdateInfo{
				VersionName: name,
				BuildNumber: uint32(version),
				Source:      "tdesktop",
			}, func() error { return s.db.SettingsStore.SetLastTDeskID(int64(version)) })
			return
		}

		if version, name, errFetch := s.tdlibVersion(); errFetch != nil {
			gologging.Error(errFetch)
		} else if int64(version) > settings.LastTdlibID {
			s.dispatch(storeTypes.UpdateInfo{
				VersionName: name,
				BuildNumber: uint32(version),
				Source:      "tdlib",
			}, func() error { return s.db.SettingsStore.SetLastTDLibID(int64(version)) })
			return
		}
	}

	info, err := storeapi.GetAppInfo()
	if err != nil {
		gologging.Error(err)
		return
	}
	if int64(info.VersionCode) <= settings.LastVersionCode && !s.patch.Load() {
		return
	}
	if err = storeapi.DownloadApk(s.cfg, info); err != nil {
		gologging.Error(err)
		return
	}
	s.dispatch(storeTypes.UpdateInfo{
		VersionName: info.Version,
		BuildNumber: info.VersionCode,
		Source:      "android",
	}, func() error { return s.db.SettingsStore.SetLastVersionCode(int64(info.VersionCode)) })
}

func (s *Service) dispatch(update storeTypes.UpdateInfo, commit func() error) {
	s.building.Store(true)
	defer s.building.Store(false)
	defer func() {
		if recovered := recover(); recovered != nil {
			gologging.Error(fmt.Sprintf("extraction panic (%s): %v\n%s", update.Source, recovered, debug.Stack()))
		}
	}()
	if err := s.extract(update); err != nil {
		gologging.Error(err)
		return
	}
	if err := commit(); err != nil {
		gologging.Error(err)
		return
	}
	s.patch.Store(false)
}

func (s *Service) tdesktopVersion(branch string) (int, string, error) {
	res, err := http.ExecuteRequest(fmt.Sprintf(consts.TDesktopSources+"/core/version.h", branch))
	if err != nil {
		return 0, "", err
	}
	body := res.String()
	code := consts.TDeskVersionRgx.FindAllStringSubmatch(body, -1)
	name := consts.TDeskVersionNameRgx.FindAllStringSubmatch(body, -1)
	if len(code) == 0 || len(name) == 0 {
		return 0, "", fmt.Errorf("tdesktop version not found")
	}
	version := 0
	_, _ = fmt.Sscanf(code[0][1], "%d", &version)
	return version, name[0][1], nil
}

func (s *Service) tdlibVersion() (int, string, error) {
	res, err := http.ExecuteRequest(consts.TDLibSources + "/CMakeLists.txt")
	if err != nil {
		return 0, "", err
	}
	name := consts.TDLibVersionRgx.FindAllStringSubmatch(res.String(), -1)
	if len(name) == 0 {
		return 0, "", fmt.Errorf("tdlib version not found")
	}
	return int(utils.VersionToCode(name[0][1])), name[0][1], nil
}

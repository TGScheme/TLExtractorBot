package services

import (
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"strconv"

	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/android"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
	"github.com/TGScheme/TLExtractorBot/internal/utils"
)

func (s *Service) pollSources() {
	if s.building.Load() {
		return
	}
	if !s.polling.CompareAndSwap(false, true) {
		return
	}
	defer s.polling.Store(false)
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
		if s.pollSource("tdesktop", settings.LastTdeskID, func() (int, string, error) {
			return s.tdesktopVersion(settings.TdesktopBranch)
		}, s.db.SettingsStore.SetLastTDeskID) {
			return
		}
		if s.pollSource("tdlib", settings.LastTdlibID, s.tdlibVersion, s.db.SettingsStore.SetLastTDLibID) {
			return
		}
	}

	post, err := s.betaPost(settings.LastPostID)
	if err != nil {
		gologging.Error(err)
		return
	}
	if post == nil {
		return
	}
	if post.Document == nil {
		if err = s.db.SettingsStore.SetLastPostID(int64(post.ID)); err != nil {
			gologging.Error(err)
		}
		return
	}
	isPatch := s.patch.Load()
	update := UpdateInfo{Source: "android"}
	if version := consts.BetaPostVersionRgx.FindStringSubmatch(post.Text); version != nil {
		build, _ := strconv.ParseUint(version[2], 10, 32)
		update.VersionName, update.BuildNumber = version[1], uint32(build)
	}
	if err = s.updateStatus(update, isPatch, stageDownloading, 0); err != nil {
		gologging.Error(err)
		return
	}
	apkPath := path.Join(s.cfg.WorkDir, consts.TempApk)
	if err = os.MkdirAll(path.Join(s.cfg.WorkDir, consts.TempBins), os.ModePerm); err != nil && !os.IsExist(err) {
		gologging.Error(err)
		return
	}
	if err = s.bot.DownloadDocument(post.Document, apkPath, func(percentage int64) {
		_ = s.updateStatus(update, isPatch, stageDownloading, percentage)
	}); err != nil {
		gologging.Error(err)
		if errStatus := s.bot.DropStatus(); errStatus != nil {
			gologging.Error(errStatus)
		}
		return
	}
	info, err := android.ReadAPKInfo(apkPath)
	if err != nil {
		gologging.Error(err)
		return
	}
	buildNumber := info.VersionCode / 10
	if int64(buildNumber) <= settings.LastVersionCode && !isPatch {
		gologging.Info(fmt.Sprintf(
			"beta channel: post %d carries %s (%d), not newer than %d",
			post.ID, info.VersionName, buildNumber, settings.LastVersionCode,
		))
		if err = s.db.SettingsStore.SetLastPostID(int64(post.ID)); err != nil {
			gologging.Error(err)
		}
		if err = s.bot.DropStatus(); err != nil {
			gologging.Error(err)
		}
		return
	}
	if err = s.db.SettingsStore.SetLastPostID(int64(post.ID)); err != nil {
		gologging.Error(err)
		return
	}
	update.VersionName, update.BuildNumber = info.VersionName, buildNumber
	s.dispatch(update, func() error {
		return s.db.SettingsStore.SetLastVersionCode(int64(buildNumber))
	})
}

func (s *Service) pollSource(
	source string,
	last int64,
	fetch func() (int, string, error),
	commit func(int64) error,
) bool {
	version, name, err := fetch()
	if err != nil {
		gologging.Error(err)
		return false
	}
	if int64(version) <= last {
		return false
	}
	s.dispatch(UpdateInfo{
		VersionName: name,
		BuildNumber: uint32(version),
		Source:      source,
	}, func() error { return commit(int64(version)) })
	return true
}

func (s *Service) betaPost(lastPostID int64) (*bot.ChannelPost, error) {
	if s.patch.Load() {
		if lastPostID > 0 {
			return s.bot.GetChannelPost(consts.AndroidBetaChannel, int(lastPostID))
		}
		latest, err := s.bot.LatestChannelPost(consts.AndroidBetaChannel)
		if err != nil {
			return nil, err
		}
		return s.bot.GetChannelPost(consts.AndroidBetaChannel, latest)
	}
	if lastPostID == 0 {
		latest, err := s.bot.LatestChannelPost(consts.AndroidBetaChannel)
		if err != nil {
			return nil, err
		}
		gologging.Info(fmt.Sprintf("beta channel: starting from post %d", latest))
		return nil, s.db.SettingsStore.SetLastPostID(int64(latest))
	}
	return s.bot.NextChannelPost(consts.AndroidBetaChannel, int(lastPostID))
}

func (s *Service) dispatch(update UpdateInfo, commit func() error) {
	s.building.Store(true)
	defer s.building.Store(false)
	defer s.patch.Store(false)
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
	}
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

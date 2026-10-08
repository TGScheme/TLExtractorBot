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
	"github.com/gotd/td/tg"
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

	if s.patch.Load() {
		if channel, post := s.patchPost(); post != nil {
			s.processChannelPost(channel, post, settings.LastVersionCode, true)
		}
		return
	}
	for _, channel := range consts.AndroidChannels {
		if s.pollChannel(channel, settings.LastVersionCode) {
			return
		}
	}
}

func (s *Service) pollChannel(channel string, lastVersion int64) bool {
	cursor, err := s.db.ChannelsStore.GetChannelCursor(channel)
	if err != nil {
		gologging.Error(err)
		return false
	}
	if cursor == 0 {
		latest, errLatest := s.bot.LatestChannelPost(channel)
		if errLatest != nil {
			gologging.Error(errLatest)
			return false
		}
		gologging.Info(fmt.Sprintf("%s: starting from post %d", channel, latest))
		if err = s.db.ChannelsStore.SetChannelCursor(channel, int64(latest)); err != nil {
			gologging.Error(err)
		}
		return false
	}
	post, err := s.bot.NextChannelPost(channel, int(cursor))
	if err != nil {
		gologging.Error(err)
		return false
	}
	if post == nil {
		return false
	}
	return s.processChannelPost(channel, post, lastVersion, false)
}

func (s *Service) patchPost() (string, *bot.ChannelPost) {
	var bestChannel string
	var best *bot.ChannelPost
	var bestBuild uint64
	for _, channel := range consts.AndroidChannels {
		cursor, err := s.db.ChannelsStore.GetChannelCursor(channel)
		if err != nil {
			gologging.Error(err)
			continue
		}
		if cursor == 0 {
			latest, errLatest := s.bot.LatestChannelPost(channel)
			if errLatest != nil {
				gologging.Error(errLatest)
				continue
			}
			cursor = int64(latest)
		}
		post, err := s.bot.GetChannelPost(channel, int(cursor))
		if err != nil {
			gologging.Error(err)
			continue
		}
		if post == nil || post.Document == nil {
			continue
		}
		if build := postBuild(post.Text); best == nil || build > bestBuild {
			bestChannel, best, bestBuild = channel, post, build
		}
	}
	return bestChannel, best
}

func postBuild(text string) uint64 {
	version := consts.BetaPostVersionRgx.FindStringSubmatch(text)
	if version == nil {
		return 0
	}
	build, _ := strconv.ParseUint(version[2], 10, 32)
	return build
}

func (s *Service) processChannelPost(channel string, post *bot.ChannelPost, lastVersion int64, isPatch bool) bool {
	advance := func() bool {
		if err := s.db.ChannelsStore.SetChannelCursor(channel, int64(post.ID)); err != nil {
			gologging.Error(err)
			return false
		}
		return true
	}
	if post.Document == nil {
		advance()
		return false
	}
	update := UpdateInfo{Source: "android"}
	if version := consts.BetaPostVersionRgx.FindStringSubmatch(post.Text); version != nil {
		update.VersionName, update.BuildNumber = version[1], uint32(postBuild(post.Text))
	}
	if !isPatch && update.BuildNumber != 0 && int64(update.BuildNumber) <= lastVersion {
		gologging.Info(fmt.Sprintf(
			"%s: post %d carries %s (%d), already extracted up to %d",
			channel, post.ID, update.VersionName, update.BuildNumber, lastVersion,
		))
		advance()
		return false
	}
	info, err := s.downloadApk(update, isPatch, post.Document)
	if err != nil {
		gologging.Error(err)
		return true
	}
	buildNumber := info.VersionCode / 10
	if int64(buildNumber) <= lastVersion && !isPatch {
		gologging.Info(fmt.Sprintf(
			"%s: post %d carries %s (%d), not newer than %d",
			channel, post.ID, info.VersionName, buildNumber, lastVersion,
		))
		advance()
		if err = s.bot.DropStatus(); err != nil {
			gologging.Error(err)
		}
		return true
	}
	if !advance() {
		return true
	}
	update.VersionName, update.BuildNumber = info.VersionName, buildNumber
	_ = s.dispatch(update, func() error {
		return s.db.SettingsStore.SetLastVersionCode(int64(buildNumber))
	})
	return true
}

func (s *Service) downloadApk(update UpdateInfo, isPatch bool, document *tg.Document) (*android.APKInfo, error) {
	s.updateStatus(update, isPatch, stageDownloading, 0)
	info, err := s.fetchApk(document, func(percentage int64) {
		s.updateStatus(update, isPatch, stageDownloading, percentage)
	})
	if err != nil {
		if errStatus := s.bot.DropStatus(); errStatus != nil {
			gologging.Error(errStatus)
		}
		return nil, err
	}
	return info, nil
}

func (s *Service) fetchApk(document *tg.Document, onProgress func(percentage int64)) (*android.APKInfo, error) {
	apkPath := path.Join(s.cfg.WorkDir, consts.TempApk)
	if err := os.MkdirAll(path.Join(s.cfg.WorkDir, consts.TempBins), os.ModePerm); err != nil {
		return nil, err
	}
	if err := s.bot.DownloadDocument(document, apkPath, onProgress); err != nil {
		return nil, err
	}
	return android.ReadAPKInfo(apkPath)
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
	_ = s.dispatch(UpdateInfo{
		VersionName: name,
		BuildNumber: uint32(version),
		Source:      source,
	}, func() error { return commit(int64(version)) })
	return true
}

func (s *Service) dispatch(update UpdateInfo, commit func() error) (err error) {
	s.building.Store(true)
	defer s.building.Store(false)
	defer s.patch.Store(false)
	defer func() {
		if recovered := recover(); recovered != nil {
			gologging.Error(fmt.Sprintf("extraction panic (%s): %v\n%s", update.Source, recovered, debug.Stack()))
			err = fmt.Errorf("extraction panic: %v", recovered)
		}
	}()
	if err = s.extract(update); err != nil {
		gologging.Error(err)
		return err
	}
	if err = commit(); err != nil {
		gologging.Error(err)
	}
	return err
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

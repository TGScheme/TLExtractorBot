package services

import (
	"fmt"

	"github.com/Laky-64/gologging"
)

func (s *Service) ExtractUploadedApk(messageID int, force bool, notify func(text string)) {
	if s.building.Load() || !s.polling.CompareAndSwap(false, true) {
		notify("<b>⏳ An extraction is already running</b>, send the apk again once it is over.")
		return
	}
	defer s.polling.Store(false)
	if force {
		s.patch.Store(true)
		defer s.patch.Store(false)
	}
	settings, err := s.db.SettingsStore.GetSettings()
	if err != nil {
		notify(fmt.Sprintf("Cannot read settings: %v", err))
		return
	}
	document, err := s.bot.ChatDocument(messageID)
	if err != nil {
		notify(fmt.Sprintf("Cannot read the apk: %v", err))
		return
	}
	isPatch := s.patch.Load()
	update := UpdateInfo{Source: "android"}
	info, err := s.downloadApk(update, isPatch, document)
	if err != nil {
		gologging.Error(err)
		notify(fmt.Sprintf("Cannot download the apk: %v", err))
		return
	}
	buildNumber := info.VersionCode / 10
	if int64(buildNumber) <= settings.LastVersionCode && !isPatch {
		if err = s.bot.DropStatus(); err != nil {
			gologging.Error(err)
		}
		notify(fmt.Sprintf(
			"<b>%s (%d)</b> is not newer than build %d, send it again with /patch as caption to extract it anyway.",
			info.VersionName, buildNumber, settings.LastVersionCode,
		))
		return
	}
	update.VersionName, update.BuildNumber = info.VersionName, buildNumber
	notify(fmt.Sprintf("<b>📥 Extracting %s</b>", update.Display()))
	if err = s.dispatch(update, func() error {
		return s.db.SettingsStore.SetLastVersionCode(int64(buildNumber))
	}); err != nil {
		notify(fmt.Sprintf("<b>❌ Extraction failed</b>\n<code>%v</code>", err))
		return
	}
	notify(fmt.Sprintf("<b>✅ %s extracted</b>", update.Display()))
}

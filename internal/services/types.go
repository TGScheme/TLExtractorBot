package services

import "fmt"

type UpdateInfo struct {
	VersionName string
	BuildNumber uint32
	Source      string
}

func (u UpdateInfo) Display() string {
	name := map[string]string{
		"tdesktop": "Telegram Desktop",
		"android":  "Telegram for Android",
		"ios":      "Telegram for iOS",
		"tdlib":    "TDLib",
	}[u.Source]
	if u.Source == "android" || u.Source == "ios" {
		return fmt.Sprintf("%s %s (%d)", name, u.VersionName, u.BuildNumber)
	}
	return fmt.Sprintf("%s %s", name, u.VersionName)
}

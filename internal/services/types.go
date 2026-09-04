package services

import "github.com/TGScheme/TLExtractorBot/internal/assets"

type UpdateInfo struct {
	VersionName string
	BuildNumber uint32
	Source      string
}

func (u UpdateInfo) Display() string {
	return assets.Render("os_full", map[string]any{"update": u})
}

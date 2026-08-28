package services

import (
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	storeTypes "github.com/TGScheme/TLExtractorBot/internal/storeapi/types"
)

const (
	stageDownloading = 0
	stageDecompiling = 1
	stageExtracting  = 2
	stagePublishing  = 3
)

func statusArgs(update storeTypes.UpdateInfo, isPatch bool, stage int, progress int64) map[string]any {
	return map[string]any{
		"update": update, "is_patch": isPatch, "stage": stage, "progress": progress,
	}
}

func (s *Service) updateStatus(update storeTypes.UpdateInfo, isPatch bool, stage int, progress int64) error {
	args := statusArgs(update, isPatch, stage, progress)
	return s.bot.UpdateRichStatus(
		assets.Render("status_message", args),
		assets.Render("message", args),
	)
}

func initialStage(source string) int {
	if source == "android" {
		return stageDownloading
	}
	return stageExtracting
}

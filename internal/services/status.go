package services

import (
	"github.com/TGScheme/TLExtractorBot/internal/assets"
)

const (
	stageDownloading = 0
	stageDecompiling = 1
	stageExtracting  = 2
	stagePublishing  = 3
)

func statusArgs(update UpdateInfo, isPatch bool, stage int, progress int64) map[string]any {
	return map[string]any{
		"update": update, "is_patch": isPatch && stage == stagePublishing, "stage": stage, "progress": progress,
	}
}

func (s *Service) updateStatus(update UpdateInfo, isPatch bool, stage int, progress int64) {
	s.bot.UpdateRichStatus(assets.Render("status_message", statusArgs(update, isPatch, stage, progress)))
}

func (s *Service) updateGeminiStatus(update UpdateInfo, isPatch bool) {
	args := statusArgs(update, isPatch, stagePublishing, 100)
	args["gemini_busy"] = true
	s.bot.UpdateRichStatus(assets.Render("status_message", args))
}

func initialStage(source string) int {
	if source == "android" {
		return stageDownloading
	}
	return stageExtracting
}

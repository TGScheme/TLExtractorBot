package services

import (
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
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
	s.bot.UpdateRichStatus(assets.Render("status_message", statusArgs(update, isPatch, stage, progress)), nil)
}

func (s *Service) updateAIStatus(update UpdateInfo, isPatch bool, draftURL string) {
	args := statusArgs(update, isPatch, stagePublishing, 100)
	args["ai_busy"] = true
	args["ai_provider"] = s.llm.ActiveKind()
	args["has_draft"] = draftURL != ""
	var button *bot.StatusButton
	if draftURL != "" {
		button = &bot.StatusButton{Text: "Technical Changes", URL: draftURL}
	}
	s.bot.UpdateRichStatus(assets.Render("status_message", args), button)
}

func initialStage(source string) int {
	if source == "android" {
		return stageDownloading
	}
	return stageExtracting
}

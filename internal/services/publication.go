package services

import (
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/gemini"
	githubTypes "github.com/TGScheme/TLExtractorBot/internal/github/types"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

type publication struct {
	update        UpdateInfo
	fullScheme    *schemeTypes.TLFullScheme
	differences   *schemeTypes.TLFullDifferences
	stableDiffs   *schemeTypes.TLFullDifferences
	previousLayer int
	incremental   bool
	isPatch       bool
	stableRelease bool
	firstStable   bool
	commit        *githubTypes.CommitInfo
	changelog     *gemini.Changelog
	aiModel       string
	title         string
	pageBanner    string
	hasBanner     bool
}

func (p publication) stats() schemeTypes.DifferenceStats {
	if p.stableRelease {
		return scheme.GetStats(p.stableDiffs)
	}
	return scheme.GetStats(p.differences)
}

func (p publication) bannerStats() schemeTypes.DifferenceStats {
	if p.firstStable {
		return scheme.GetStats(p.stableDiffs)
	}
	return scheme.GetStats(p.differences)
}

func (p publication) page() string {
	stableStats := scheme.GetStats(p.stableDiffs)
	args := map[string]any{
		"layer":               p.fullScheme.Layer,
		"previous_layer":      p.previousLayer,
		"differences":         p.stableDiffs,
		"stats":               stableStats,
		"commit_urls":         p.commit.FilesLines,
		"banner_url":          p.pageBanner,
		"main_scheme":         scheme.ToString(p.stableDiffs.MainApi, p.fullScheme.Layer, false),
		"e2e_scheme":          scheme.ToString(p.stableDiffs.E2EApi, p.fullScheme.Layer, false),
		"gemini_descriptions": map[string]string{},
		"update":              p.update,
		"is_stable":           p.fullScheme.IsSync,
	}
	if p.incremental && !p.stableRelease {
		args["latest"] = p.differences
		args["latest_stats"] = scheme.GetStats(p.differences)
		args["latest_source"] = p.update
	}
	if p.changelog != nil {
		args["story"] = p.changelog
		args["gemini_descriptions"] = p.changelog.Descriptions
		args["ai_model"] = p.aiModel
	}
	return assets.Render("changelogs", args)
}

func (p publication) post() string {
	stats := p.stats()
	changes := richChanges(p.differences)
	lead := ""
	if p.changelog != nil {
		lead = p.changelog.Lead
	}
	args := map[string]any{
		"update": p.update, "layer": p.fullScheme.Layer, "stats": stats,
		"is_stable": p.fullScheme.IsSync, "is_patch": p.isPatch,
		"is_stable_release": p.stableRelease, "previous_layer": p.previousLayer,
		"title": p.title, "lead": lead, "changes": changes,
		"total_stats":    combinedStats(stats),
		"patch_summary":  patchSummary(len(changes)),
		"commit_urls":    p.commit.FilesLines,
		"is_incremental": p.incremental,
	}
	if p.hasBanner {
		args["banner_url"] = bot.BannerSource
	}
	return assets.Render("rich_message", args)
}

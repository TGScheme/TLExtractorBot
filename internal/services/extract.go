package services

import (
	"errors"
	"fmt"
	"slices"

	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/android"
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/banner"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/db/models"
	"github.com/TGScheme/TLExtractorBot/internal/gemini"
	"github.com/TGScheme/TLExtractorBot/internal/java/jadx"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	telegraphTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
)

const maxReportedProblems = 10

func (s *Service) extract(update UpdateInfo) error {
	isPatch := s.patch.Load()
	stage := initialStage(update.Source)
	if update.Source == "android" {
		stage = stageDecompiling
	}
	if err := s.updateStatus(update, isPatch, stage, 0); err != nil {
		return err
	}

	settings, err := s.db.SettingsStore.GetSettings()
	if err != nil {
		return err
	}
	preview, err := s.db.LoadScheme(models.SchemeRoleEnumPreview)
	if err != nil {
		return err
	}
	previewLayer := 0
	if preview != nil {
		previewLayer = preview.Layer
	}

	fullScheme, err := s.buildScheme(update, settings.TdesktopBranch, preview, previewLayer, isPatch)
	if err != nil {
		return err
	}

	if fixed := scheme.FixNamespaces(fullScheme); fixed > 0 {
		gologging.Info(fmt.Sprintf("scheme: fixed the namespace of %d objects", fixed))
	}
	if fixed := scheme.FixFieldNames(fullScheme); fixed > 0 {
		gologging.Info(fmt.Sprintf("scheme: fixed the name of %d fields", fixed))
	}
	if applied := scheme.ApplyOverrides(fullScheme); applied > 0 {
		gologging.Info(fmt.Sprintf("scheme: applied %d overrides", applied))
	}

	differences := scheme.GetDiffs(preview, fullScheme)
	if differences == nil || fullScheme.Layer < previewLayer {
		return s.bot.DropStatus()
	}
	return s.publish(update, fullScheme, preview, differences, isPatch)
}

func (s *Service) buildScheme(
	update UpdateInfo,
	branch string,
	preview *schemeTypes.TLFullScheme,
	previewLayer int,
	isPatch bool,
) (*schemeTypes.TLFullScheme, error) {
	if update.Source == "android" {
		_ = s.updateStatus(update, isPatch, stageDecompiling, 0)
		if err := jadx.Decompile(s.cfg, func(percentage int64) {
			_ = s.updateStatus(update, isPatch, stageDecompiling, percentage)
		}); err != nil {
			return nil, err
		}
		_ = s.updateStatus(update, isPatch, stageExtracting, 100)
		fullScheme, extracted, err := android.ExtractScheme(s.cfg.WorkDir, s.scheme, branch)
		if err != nil {
			return nil, err
		}
		if err = s.applyDisappeared(update.Source, fullScheme, extracted); err != nil {
			return nil, err
		}
		return fullScheme, nil
	}

	var remoteScheme *schemeTypes.TLRemoteScheme
	var patchOs schemeTypes.PatchOS
	var err error
	switch update.Source {
	case "tdesktop":
		remoteScheme, err = scheme.GetScheme(branch)
		patchOs = schemeTypes.TDesktopPatch
	case "tdlib":
		remoteScheme, err = scheme.GetTDLibScheme()
		patchOs = schemeTypes.TDLibPatch
	default:
		return nil, errors.New("unknown source")
	}
	if err != nil {
		return nil, err
	}
	if err = s.scheme.UpdateUpstreamCache(update.Source, remoteScheme, branch); err != nil {
		return nil, err
	}
	if preview == nil {
		return nil, errors.New("no preview scheme to merge against")
	}
	fullScheme, err := s.scheme.MergeRemote(
		remoteScheme, patchOs, remoteScheme.Layer == previewLayer, true,
		func(isE2E bool) (*schemeTypes.TLRemoteScheme, error) {
			return clonePreview(preview, previewLayer, isE2E), nil
		},
	)
	if err != nil {
		return nil, err
	}
	fullScheme.IsSync = true
	return fullScheme, nil
}

func clonePreview(preview *schemeTypes.TLFullScheme, layer int, isE2E bool) *schemeTypes.TLRemoteScheme {
	source := preview.MainApi
	if isE2E {
		source = preview.E2EApi
	}
	cloned := &schemeTypes.TLRemoteScheme{Layer: layer}
	for _, method := range source.Methods {
		cloned.Methods = append(cloned.Methods, &schemeTypes.TLMethod{
			TLBase: method.TLBase.Clone(),
			Method: method.Method,
		})
	}
	for _, constructor := range source.Constructors {
		cloned.Constructors = append(cloned.Constructors, &schemeTypes.TLConstructor{
			TLBase:    constructor.TLBase.Clone(),
			Predicate: constructor.Predicate,
		})
	}
	return cloned
}

func (s *Service) publish(
	update UpdateInfo,
	fullScheme *schemeTypes.TLFullScheme,
	preview *schemeTypes.TLFullScheme,
	differences *schemeTypes.TLFullDifferences,
	isPatch bool,
) error {
	problems := scheme.Validate(fullScheme)
	if fatal := scheme.FatalProblems(problems); len(fatal) > 0 {
		return s.reportProblems(update, fullScheme.Layer, fatal, true)
	}
	if len(problems) > 0 {
		if err := s.reportProblems(update, fullScheme.Layer, problems, false); err != nil {
			gologging.Error("scheme: unable to report the validation warnings:", err)
		}
	}

	_ = s.updateStatus(update, isPatch, stagePublishing, 100)

	stats := scheme.GetStats(differences)
	commitMessage := fmt.Sprintf("Updated to Layer %d", fullScheme.Layer)
	if isPatch {
		commitMessage = fmt.Sprintf("Patch %d", fullScheme.Layer)
	}
	commitInfo, err := s.github.MakeCommit(fullScheme, stats, commitMessage)
	if err != nil {
		return err
	}

	stableScheme, err := s.stableBaseline(preview, fullScheme.Layer)
	if err != nil {
		return err
	}
	stableDiffs := scheme.GetDiffs(stableScheme, fullScheme)
	if stableDiffs == nil {
		stableDiffs = differences
	}
	stableStats := scheme.GetStats(stableDiffs)

	previousLayer := 0
	if stableScheme != nil && stableScheme.Layer != fullScheme.Layer {
		previousLayer = stableScheme.Layer
	}
	pageArgs := map[string]any{
		"layer":               fullScheme.Layer,
		"previous_layer":      previousLayer,
		"differences":         stableDiffs,
		"stats":               stableStats,
		"commit_urls":         commitInfo.FilesLines,
		"banner_url":          s.cfg.BannerURL,
		"main_scheme":         scheme.ToString(stableDiffs.MainApi, fullScheme.Layer, false),
		"e2e_scheme":          scheme.ToString(stableDiffs.E2EApi, fullScheme.Layer, false),
		"gemini_descriptions": map[string]string{},
		"update":              update,
		"is_stable":           fullScheme.IsSync,
	}
	if preview != nil && preview.Layer == fullScheme.Layer {
		pageArgs["latest"] = differences
		pageArgs["latest_stats"] = stats
		pageArgs["latest_source"] = update
	}
	lead, title := "", ""
	changelog, err := s.gemini.GenerateChangelog(gemini.ChangelogRequest{
		Layer:       fullScheme.Layer,
		Source:      update.Source,
		VersionName: update.VersionName,
		BuildNumber: update.BuildNumber,
		IsPatch:     isPatch,
		Scheme:      fullScheme,
		Differences: stableDiffs,
	})
	if err != nil {
		gologging.Error("gemini: unable to generate the changelog:", err)
	} else if changelog != nil {
		pageArgs["story"] = changelog
		pageArgs["gemini_descriptions"] = changelog.Descriptions
		pageArgs["ai_model"] = s.gemini.Model()
		lead, title = changelog.Lead, changelog.Title
	}

	bannerURL, pageBannerURL := s.layerBanner(update, fullScheme, stats, title, isPatch)
	pageArgs["banner_url"] = pageBannerURL

	pageTitle := fmt.Sprintf("Layer %d", fullScheme.Layer)
	if title != "" {
		pageTitle = title
	}

	url, err := s.publishPage(fullScheme.Layer, pageTitle, assets.Render("changelogs", pageArgs))
	if err != nil {
		return err
	}
	keyboard := &tgTypes.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgTypes.InlineKeyboardButton{{
			{Text: "Full Changelog", URL: url},
			{Text: "GitHub", URL: commitInfo.SourceURL},
		}},
	}
	richArgs := map[string]any{
		"update": update, "layer": fullScheme.Layer, "stats": stats,
		"is_stable": fullScheme.IsSync, "is_patch": isPatch,
	}
	richArgs["title"] = title
	richArgs["lead"] = lead
	changes := richChanges(differences)
	richArgs["changes"] = changes
	richArgs["total_stats"] = combinedStats(stats)
	richArgs["patch_summary"] = patchSummary(len(changes))
	richArgs["commit_urls"] = commitInfo.FilesLines
	richArgs["is_incremental"] = preview != nil && preview.Layer == fullScheme.Layer
	richArgs["banner_url"] = bannerURL

	if err = s.bot.PublishRich(assets.Render("rich_message", richArgs), true, keyboard); err != nil {
		return err
	}
	return s.promote(fullScheme, preview)
}

func (s *Service) layerBanner(
	update UpdateInfo,
	fullScheme *schemeTypes.TLFullScheme,
	stats schemeTypes.DifferenceStats,
	title string,
	isPatch bool,
) (string, string) {
	if title == "" {
		title = update.Display()
	}
	totals := combinedStats(stats)
	input := banner.Input{
		Layer:        fullScheme.Layer,
		Title:        title,
		Source:       update.Display(),
		ChangesLabel: "CONSTRUCTORS",
		Highlight:    fmt.Sprintf("%d", totals.TotalAdditions),
		Changes: fmt.Sprintf(
			" added · %d changed · %d removed", totals.TotalChanges, totals.TotalDeletions,
		),
		IsStable: fullScheme.IsSync,
	}
	name := fmt.Sprintf("layer-%d", fullScheme.Layer)
	if isPatch {
		name = fmt.Sprintf("layer-%d-patch", fullScheme.Layer)
	}
	message := fmt.Sprintf("Banner for Layer %d", fullScheme.Layer)
	full, err := s.upload(name+".png", message, func() ([]byte, error) { return banner.Render(input) })
	if err != nil {
		gologging.Error("banner: unable to publish the layer banner:", err)
		return s.cfg.BannerURL, s.cfg.BannerURL
	}
	compact, err := s.upload(name+".jpg", message, func() ([]byte, error) { return banner.RenderCompact(input) })
	if err != nil {
		gologging.Error("banner: unable to publish the compact banner:", err)
		return full, full
	}
	return full, compact
}

func (s *Service) upload(name, message string, render func() ([]byte, error)) (string, error) {
	image, err := render()
	if err != nil {
		return "", err
	}
	return s.github.CommitBanner(name, image, message)
}

func (s *Service) reportProblems(update UpdateInfo, layer int, problems []scheme.Problem, blocking bool) error {
	for _, problem := range problems {
		gologging.Error("scheme:", problem)
	}
	shown, truncated := problems, 0
	if len(shown) > maxReportedProblems {
		truncated = len(shown) - maxReportedProblems
		shown = shown[:maxReportedProblems]
	}
	if blocking {
		if err := s.bot.DropStatus(); err != nil {
			return err
		}
	}
	return s.bot.LogMessage(assets.Render("scheme_problems", map[string]any{
		"layer":     layer,
		"source":    update.Source,
		"total":     len(problems),
		"problems":  shown,
		"truncated": truncated,
		"blocking":  blocking,
	}))
}

func (s *Service) publishPage(layer int, title, html string) (string, error) {
	storedPath, err := s.db.PagesStore.GetLayerPagePath(int64(layer))
	if err != nil {
		return "", err
	}
	var page telegraphTypes.PageInfo
	if storedPath != "" {
		if published, errTitle := s.telegraph.PageTitle(storedPath); errTitle != nil {
			gologging.Error("telegraph: unable to read the published title:", errTitle)
		} else if published != "" {
			title = published
		}
		if page, err = s.telegraph.EditPage(storedPath, title, html); err != nil {
			gologging.Error("telegraph: unable to edit the page, creating a new one:", err)
			storedPath = ""
		}
	}
	if storedPath == "" {
		if page, err = s.telegraph.CreatePage(title, html); err != nil {
			return "", err
		}
	}
	if err = s.db.PagesStore.SetLayerPagePath(int64(layer), page.Path); err != nil {
		return "", err
	}
	return page.URL, nil
}

func (s *Service) stableBaseline(preview *schemeTypes.TLFullScheme, layer int) (*schemeTypes.TLFullScheme, error) {
	recent, err := s.db.RecentStore.ListRecentLayers()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(recent, int64(layer)) {
		return preview, nil
	}
	stable, err := s.db.LoadScheme(models.SchemeRoleEnumStable)
	if err != nil {
		return nil, err
	}
	if stable == nil {
		return preview, nil
	}
	return stable, nil
}

func (s *Service) promote(fullScheme, preview *schemeTypes.TLFullScheme) error {
	if err := db.IgnoreNoRows(s.db.RecentStore.AddRecentLayer(int64(fullScheme.Layer))); err != nil {
		return err
	}
	recent, err := s.db.RecentStore.ListRecentLayers()
	if err != nil {
		return err
	}
	if len(recent) > 1 && preview != nil {
		if err = s.db.SaveScheme(preview, models.SchemeRoleEnumStable); err != nil {
			return err
		}
		if err = db.IgnoreNoRows(s.db.RecentStore.TrimRecentLayers(1)); err != nil {
			return err
		}
	}
	return s.db.SaveScheme(fullScheme, models.SchemeRoleEnumPreview)
}

func (s *Service) hasPreview() (bool, error) {
	preview, err := s.db.LoadScheme(models.SchemeRoleEnumPreview)
	if err != nil {
		return false, err
	}
	return preview != nil, nil
}

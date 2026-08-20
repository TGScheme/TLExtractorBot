package services

import (
	"errors"
	"fmt"
	"slices"

	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/android"
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/db/models"
	"github.com/TGScheme/TLExtractorBot/internal/gemini"
	"github.com/TGScheme/TLExtractorBot/internal/java/jadx"
	storeTypes "github.com/TGScheme/TLExtractorBot/internal/storeapi/types"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	telegraphTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
)

func (s *Service) extract(update storeTypes.UpdateInfo) error {
	isPatch := s.patch.Load()
	if err := s.bot.UpdateStatus(assets.Render("message", map[string]any{
		"update": update, "progress": 0, "is_patch": isPatch,
	}), false, false, nil); err != nil {
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

	differences := scheme.GetDiffs(preview, fullScheme)
	if differences == nil || fullScheme.Layer < previewLayer {
		return s.bot.UpdateStatus("", false, false, nil)
	}
	return s.publish(update, fullScheme, preview, differences, isPatch)
}

func (s *Service) buildScheme(
	update storeTypes.UpdateInfo,
	branch string,
	preview *schemeTypes.TLFullScheme,
	previewLayer int,
	isPatch bool,
) (*schemeTypes.TLFullScheme, error) {
	if update.Source == "android" {
		if err := jadx.Decompile(s.cfg, func(percentage int64) {
			_ = s.bot.UpdateStatus(assets.Render("message", map[string]any{
				"update": update, "progress": percentage, "is_patch": isPatch,
			}), false, false, nil)
		}); err != nil {
			return nil, err
		}
		return android.ExtractScheme(s.cfg.WorkDir, s.scheme, branch)
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
	update storeTypes.UpdateInfo,
	fullScheme *schemeTypes.TLFullScheme,
	preview *schemeTypes.TLFullScheme,
	differences *schemeTypes.TLFullDifferences,
	isPatch bool,
) error {
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

	pageTitle := fmt.Sprintf("Layer %d", fullScheme.Layer)
	if !fullScheme.IsSync {
		pageTitle += " Preview"
	}
	pageArgs := map[string]any{
		"differences":         stableDiffs,
		"stats":               stableStats,
		"commit_urls":         commitInfo.FilesLines,
		"banner_url":          s.cfg.BannerURL,
		"main_scheme":         scheme.ToString(stableDiffs.MainApi, fullScheme.Layer, false),
		"e2e_scheme":          scheme.ToString(stableDiffs.E2EApi, fullScheme.Layer, false),
		"gemini_descriptions": map[string]string{},
	}
	if preview != nil && preview.Layer == fullScheme.Layer {
		pageArgs["latest"] = differences
		pageArgs["latest_stats"] = stats
		pageArgs["latest_source"] = update
	}
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
	}

	url, err := s.publishPage(fullScheme.Layer, pageTitle, assets.Render("changelogs", pageArgs))
	if err != nil {
		return err
	}
	if err = s.bot.UpdateStatus(assets.Render("message", map[string]any{
		"update": update, "layer": fullScheme.Layer, "stats": stats,
		"is_stable": fullScheme.IsSync, "is_patch": isPatch,
	}), true, true, &tgTypes.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgTypes.InlineKeyboardButton{{
			{Text: "Full Changelog", URL: url},
			{Text: "GitHub", URL: commitInfo.SourceURL},
		}},
	}); err != nil {
		return err
	}
	return s.promote(fullScheme, preview)
}

func (s *Service) publishPage(layer int, title, html string) (string, error) {
	storedPath, err := s.db.PagesStore.GetLayerPagePath(int64(layer))
	if err != nil {
		return "", err
	}
	var page telegraphTypes.PageInfo
	if storedPath != "" {
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

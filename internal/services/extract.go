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
	githubTypes "github.com/TGScheme/TLExtractorBot/internal/github/types"
	"github.com/TGScheme/TLExtractorBot/internal/java/jadx"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

const maxReportedProblems = 10

func (s *Service) extract(update UpdateInfo) error {
	isPatch := s.patch.Load()
	stage := initialStage(update.Source)
	if update.Source == "android" {
		stage = stageDecompiling
	}
	s.updateStatus(update, isPatch, stage, 0)

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
	if isPatch && fullScheme.Layer > previewLayer {
		gologging.Info(fmt.Sprintf("scheme: layer %d was never published, releasing it instead of patching", fullScheme.Layer))
		isPatch = false
	}

	stable, err := s.db.LoadScheme(models.SchemeRoleEnumStable)
	if err != nil {
		return err
	}
	if fixed := scheme.FixNamespaces(fullScheme, preview, stable); fixed > 0 {
		gologging.Info(fmt.Sprintf("scheme: fixed the namespace of %d objects", fixed))
	}
	if fixed := scheme.FixFieldNames(fullScheme); fixed > 0 {
		gologging.Info(fmt.Sprintf("scheme: fixed the name of %d fields", fixed))
	}
	if applied := scheme.ApplyOverrides(fullScheme); applied > 0 {
		gologging.Info(fmt.Sprintf("scheme: applied %d overrides", applied))
	}

	differences := scheme.GetDiffs(preview, fullScheme)
	if fullScheme.Layer < previewLayer {
		return s.bot.DropStatus()
	}
	firstStable, err := s.becameStable(fullScheme, preview)
	if err != nil {
		return err
	}
	if differences == nil {
		if !firstStable {
			return s.bot.DropStatus()
		}
		gologging.Info(fmt.Sprintf("scheme: layer %d became stable, announcing it", fullScheme.Layer))
		isPatch = false
	}
	return s.publish(update, fullScheme, preview, differences, isPatch, firstStable)
}

func (s *Service) buildScheme(
	update UpdateInfo,
	branch string,
	preview *schemeTypes.TLFullScheme,
	previewLayer int,
	isPatch bool,
) (*schemeTypes.TLFullScheme, error) {
	if update.Source == "android" {
		s.updateStatus(update, isPatch, stageDecompiling, 0)
		if err := jadx.Decompile(s.cfg, func(percentage int64) {
			s.updateStatus(update, isPatch, stageDecompiling, percentage)
		}); err != nil {
			return nil, err
		}
		s.updateStatus(update, isPatch, stageExtracting, 100)
		fullScheme, extracted, err := android.ExtractScheme(s.cfg.WorkDir, s.scheme, branch)
		if err != nil {
			return nil, err
		}
		if err = s.applyDropped(update, fullScheme, preview, extracted); err != nil {
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
	firstStable bool,
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

	s.updateStatus(update, isPatch, stagePublishing, 100)

	stableRelease := differences == nil
	if stableRelease {
		differences = &schemeTypes.TLFullDifferences{}
	}
	var commitInfo *githubTypes.CommitInfo
	var err error
	if stableRelease {
		commitInfo, err = s.github.CurrentCommit(fullScheme)
	} else {
		commitMessage := fmt.Sprintf("Updated to Layer %d", fullScheme.Layer)
		if isPatch {
			commitMessage = fmt.Sprintf("Patch %d", fullScheme.Layer)
		}
		commitInfo, err = s.github.MakeCommit(fullScheme, scheme.GetStats(differences), commitMessage)
	}
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
	pub := publication{
		update:        update,
		fullScheme:    fullScheme,
		differences:   differences,
		stableDiffs:   stableDiffs,
		incremental:   preview != nil && preview.Layer == fullScheme.Layer,
		isPatch:       isPatch,
		stableRelease: stableRelease,
		firstStable:   firstStable,
		commit:        commitInfo,
		pageBanner:    s.cfg.BannerURL,
	}
	if stableScheme != nil && stableScheme.Layer != fullScheme.Layer {
		pub.previousLayer = stableScheme.Layer
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
		pub.changelog = changelog
		pub.aiModel = s.gemini.Model()
		pub.title = changelog.Title
	}
	if published := s.publishedTitle(fullScheme.Layer); published != "" {
		pub.title = published
	}

	bannerName := fmt.Sprintf("layer-%d", fullScheme.Layer)
	switch {
	case firstStable:
		bannerName += "-stable"
	case isPatch:
		bannerName += "-patch"
	}
	bannerImage, pageBannerURL := s.layerBanner(update, fullScheme, pub.bannerStats(), pub.title, bannerName)
	pub.pageBanner = pageBannerURL
	pub.hasBanner = bannerImage != nil

	pageTitle := fmt.Sprintf("Layer %d", fullScheme.Layer)
	if pub.title != "" {
		pageTitle = pub.title
	}
	url, err := s.publishPage(fullScheme.Layer, pageTitle, pub.page())
	if err != nil {
		return err
	}
	keyboard := &tgTypes.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgTypes.InlineKeyboardButton{{
			{Text: "Full Changelog", URL: url},
			{Text: "GitHub", URL: commitInfo.SourceURL},
		}},
	}
	if err = s.bot.PublishRich(pub.post(), true, keyboard, bannerImage); err != nil {
		return err
	}
	return s.promote(fullScheme, preview)
}

func (s *Service) layerBanner(
	update UpdateInfo,
	fullScheme *schemeTypes.TLFullScheme,
	stats schemeTypes.DifferenceStats,
	title string,
	name string,
) ([]byte, string) {
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
	pageURL := s.cfg.BannerURL
	if compact, err := banner.RenderCompact(input); err != nil {
		gologging.Error("banner: unable to render the compact banner:", err)
	} else if url, err := s.github.CommitBanner(name+".jpg", compact, fmt.Sprintf("Banner for Layer %d", fullScheme.Layer)); err != nil {
		gologging.Error("banner: unable to publish the compact banner:", err)
	} else {
		pageURL = url
	}
	full, err := banner.Render(input)
	if err != nil {
		gologging.Error("banner: unable to render the layer banner:", err)
		return nil, pageURL
	}
	return full, pageURL
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

func (s *Service) publishedTitle(layer int) string {
	path, err := s.db.PagesStore.GetLayerPagePath(int64(layer))
	if err != nil || path == "" {
		return ""
	}
	title, err := s.telegraph.PageTitle(path)
	if err != nil {
		gologging.Error("telegraph: unable to read the published title:", err)
		return ""
	}
	if title == fmt.Sprintf("Layer %d", layer) {
		return ""
	}
	return title
}

func (s *Service) publishPage(layer int, title, html string) (string, error) {
	page, err := s.telegraph.CreatePage(title, html)
	if err != nil {
		return "", err
	}
	storedPath, err := s.db.PagesStore.GetLayerPagePath(int64(layer))
	if err != nil {
		return "", err
	}
	if storedPath == "" {
		if err = s.db.PagesStore.SetLayerPagePath(int64(layer), page.Path); err != nil {
			return "", err
		}
	}
	return page.URL, nil
}

func (s *Service) stableBaseline(preview *schemeTypes.TLFullScheme, layer int) (*schemeTypes.TLFullScheme, error) {
	for _, role := range []models.SchemeRoleEnum{models.SchemeRoleEnumSynced, models.SchemeRoleEnumSyncedPrevious} {
		synced, err := s.db.LoadScheme(role)
		if err != nil {
			return nil, err
		}
		if synced != nil && synced.Layer < layer {
			return synced, nil
		}
	}
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

func (s *Service) syncedScheme(preview *schemeTypes.TLFullScheme) (*schemeTypes.TLFullScheme, error) {
	synced, err := s.db.LoadScheme(models.SchemeRoleEnumSynced)
	if err != nil || synced != nil || preview == nil || !preview.IsSync {
		return synced, err
	}
	if err = s.db.SaveScheme(preview, models.SchemeRoleEnumSynced); err != nil {
		return nil, err
	}
	return preview, nil
}

func (s *Service) becameStable(fullScheme, preview *schemeTypes.TLFullScheme) (bool, error) {
	if !fullScheme.IsSync {
		return false, nil
	}
	synced, err := s.syncedScheme(preview)
	if err != nil {
		return false, err
	}
	return synced == nil || synced.Layer < fullScheme.Layer, nil
}

func (s *Service) promoteSynced(fullScheme, preview *schemeTypes.TLFullScheme) error {
	if !fullScheme.IsSync {
		return nil
	}
	synced, err := s.syncedScheme(preview)
	if err != nil {
		return err
	}
	if synced != nil && synced.Layer < fullScheme.Layer {
		if err = s.db.SaveScheme(synced, models.SchemeRoleEnumSyncedPrevious); err != nil {
			return err
		}
	}
	return s.db.SaveScheme(fullScheme, models.SchemeRoleEnumSynced)
}

func (s *Service) promote(fullScheme, preview *schemeTypes.TLFullScheme) error {
	if err := s.promoteSynced(fullScheme, preview); err != nil {
		return err
	}
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

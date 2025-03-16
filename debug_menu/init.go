package debug_menu

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"TLExtractor/tui"
	tuiTypes "TLExtractor/tui/types"
	"TLExtractor/utils/package_manager"
	"fmt"
	"github.com/charmbracelet/huh"
)

var ReadyToTest bool

func init() {
	debugApp := tui.NewMiniApp("debug")
	var extractStableScheme bool
	var extractPreviewScheme bool
	debugApp.SetFields(
		huh.NewConfirm().
			Title("Do you want to extract stable scheme?").
			Description("If you want to extract stable scheme, you can enter it. Otherwise, keep current scheme.").
			Value(&extractStableScheme),
	)
	debugApp.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		if environment.Debug {
			package_manager.CheckPackages()
			return fmt.Errorf("debug mode")
		}
		return nil
	}, tuiTypes.InitCheck)

	stablePage := debugApp.NewAppPage()
	stablePage.HideFunc(func() bool {
		return !extractStableScheme
	})
	stableRelease, stableTDesk := newSelectors("stable")
	stablePage.SetFields(
		stableRelease,
		stableTDesk,
	)
	stablePage.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		if checkType == tuiTypes.SubmitCheck {
			environment.LocalStorage.RecentLayers = []int{}
			environment.LocalStorage.PatchedObjects = make(map[schemeTypes.PatchOS]map[string]*schemeTypes.PatchInfo)
			environment.LocalStorage.PatchedObjects[schemeTypes.AndroidPatch] = make(map[string]*schemeTypes.PatchInfo)
			consts.TDesktopBranch = stableTDesk.GetCommitSha()
			scheme, err := extractScheme(stablePage, "stable")
			if err != nil {
				return err
			}
			environment.LocalStorage.StableLayer = scheme
			environment.LocalStorage.Commit()
		}
		return nil
	}, tuiTypes.InitCheck, tuiTypes.SubmitCheck)

	confirmPreview := debugApp.NewAppPage()
	confirmPreview.SetFields(
		huh.NewConfirm().
			Title("Do you want to extract preview scheme?").
			Description("If you want to extract preview scheme, you can enter it. Otherwise, keep current scheme.").
			Value(&extractPreviewScheme),
	)

	previewPage := debugApp.NewAppPage()
	previewPage.HideFunc(func() bool {
		return !extractPreviewScheme
	})
	previewRelease, previewTDesk := newSelectors("preview")
	previewPage.SetFields(
		previewRelease,
		previewTDesk,
	)
	previewPage.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		consts.TDesktopBranch = previewTDesk.GetCommitSha()
		scheme, err := extractScheme(previewPage, "preview")
		if err != nil {
			return err
		}
		environment.LocalStorage.PreviewLayer = scheme
		environment.LocalStorage.Commit()
		return nil
	}, tuiTypes.SubmitCheck)

	latestPage := debugApp.NewAppPage()
	latestRelease, latestTDesk := newSelectors("latest")
	latestPage.SetFields(
		latestRelease,
		latestTDesk,
	)
	latestPage.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		consts.TDesktopBranch = latestTDesk.GetCommitSha()
		ReadyToTest = true
		return nil
	}, tuiTypes.SubmitCheck)
}

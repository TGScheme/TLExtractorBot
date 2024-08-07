package main

import (
	"TLExtractor/android"
	"TLExtractor/appcenter"
	"TLExtractor/appcenter/types"
	"TLExtractor/github"
	"TLExtractor/java/jadx"
	"TLExtractor/resources"
	"TLExtractor/screen"
	"TLExtractor/telegram/bot"
	"TLExtractor/telegram/scheme"
	"TLExtractor/telegram/telegraph"
	"TLExtractor/utils"
	"TLExtractor/utils/package_manager"
	"embed"
	"fmt"
	"slices"
	"time"
)

//go:embed templates
var templatesFolder embed.FS

func main() {
	utils.LoadShellFlags()
	if err := resources.Load(templatesFolder); err != nil {
		utils.CrashLog(err, true)
	}
	if err := utils.LoadConfigs(); err != nil {
		utils.CrashLog(err, true)
	}
	telegraphClient, err := telegraph.Login()
	if err != nil {
		utils.CrashLog(err, true)
	}
	githubClient, err := github.Login()
	if err != nil {
		utils.CrashLog(err, true)
	}
	client, err := bot.NewClient()
	if err != nil {
		utils.CrashLog(err, true)
	}
	if err = package_manager.CheckPackages(githubClient); err != nil {
		utils.CrashLog(err, true)
	}
	startingScreen, err := screen.CheckScreen()
	if err != nil {
		utils.CrashLog(err, true)
	}
	if startingScreen {
		return
	}
	appcenter.Listen(func(update types.UpdateInfo) error {
		if err = client.UpdateStatus(
			resources.Format(
				"message",
				map[string]any{
					"update":   update,
					"progress": 0,
				},
			),
			false,
			false,
		); err != nil {
			return err
		}
		startTime := time.Now()
		if err = jadx.Decompile(func(percentage int64) {
			if percentage == 100 {
				return
			}
			if err = client.UpdateStatus(
				resources.Format(
					"message",
					map[string]any{
						"update":   update,
						"progress": percentage,
					},
				),
				false,
				false,
			); err != nil {
				utils.CrashLog(err, true)
			}
		}); err != nil {
			return err
		}
		elapsedTime := time.Since(startTime)
		fullScheme, err := android.ExtractScheme()
		if err != nil {
			return err
		}
		if differences := scheme.GetDiffs(utils.LocalStorage.PreviewLayer, fullScheme); differences != nil {
			stats := scheme.GetStats(differences)
			commitUrls, err := githubClient.MakeCommit(
				fullScheme,
				stats,
				fmt.Sprintf("Updated to Layer %d", fullScheme.Layer),
			)
			if err != nil {
				return err
			}
			stableDiffs := scheme.GetDiffs(
				utils.LocalStorage.StableLayer,
				fullScheme,
			)
			url, err := telegraphClient.CreatePage(
				fmt.Sprintf("Layer %d Preview", fullScheme.Layer),
				resources.Format(
					"changelogs",
					map[string]any{
						"differences": differences,
						"stats":       stats,
						"commit_urls": commitUrls,
						"banner_url":  utils.LocalStorage.BannerURL,
						"main_scheme": scheme.ToString(stableDiffs.MainApi, fullScheme.Layer, false),
						"e2e_scheme":  scheme.ToString(stableDiffs.E2EApi, fullScheme.Layer, false),
					},
				),
			)
			if err != nil {
				return err
			}
			if err = client.UpdateStatus(
				resources.Format(
					"message",
					map[string]any{
						"update": update,
						"time":   utils.FormatDuration(elapsedTime),
						"layer":  fullScheme.Layer,
						"stats":  stats,
						"link":   url,
					},
				),
				true,
				true,
			); err != nil {
				return err
			}
		}
		if !slices.Contains(utils.ShellFlags, "debug") {
			if len(utils.LocalStorage.RecentLayers) == 0 {
				utils.LocalStorage.StableLayer = utils.LocalStorage.PreviewLayer
			}
			if !slices.Contains(utils.LocalStorage.RecentLayers, fullScheme.Layer) {
				utils.LocalStorage.RecentLayers = append(utils.LocalStorage.RecentLayers, fullScheme.Layer)
			}
			if len(utils.LocalStorage.RecentLayers) > 1 {
				utils.LocalStorage.RecentLayers = utils.LocalStorage.RecentLayers[1:]
				utils.LocalStorage.StableLayer = utils.LocalStorage.PreviewLayer
			}
			utils.LocalStorage.PreviewLayer = fullScheme
			if err = utils.LocalStorage.Commit(); err != nil {
				return err
			}
		}
		return nil
	})
}

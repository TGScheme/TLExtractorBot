package main

import (
	"TLExtractor/android"
	"TLExtractor/appcenter"
	"TLExtractor/appcenter/types"
	"TLExtractor/environment"
	"TLExtractor/github"
	"TLExtractor/java/jadx"
	"TLExtractor/logging"
	_ "TLExtractor/screen"
	"TLExtractor/telegram/bot"
	"TLExtractor/telegram/scheme"
	"TLExtractor/telegram/telegraph"
	"TLExtractor/utils"
	_ "TLExtractor/utils/package_manager"
	"fmt"
	"slices"
	"time"
)

func main() {
	appcenter.Listen(func(update types.UpdateInfo) error {
		if err := bot.Client.UpdateStatus(
			environment.FormatVar(
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
		if err := jadx.Decompile(func(percentage int64) {
			if percentage == 100 {
				return
			}
			if err := bot.Client.UpdateStatus(
				environment.FormatVar(
					"message",
					map[string]any{
						"update":   update,
						"progress": percentage,
					},
				),
				false,
				false,
			); err != nil {
				logging.Fatal(err)
			}
		}); err != nil {
			return err
		}
		elapsedTime := time.Since(startTime)
		fullScheme, err := android.ExtractScheme()
		if err != nil {
			return err
		}
		if differences := scheme.GetDiffs(environment.LocalStorage.PreviewLayer, fullScheme); differences != nil {
			stats := scheme.GetStats(differences)
			commitUrls, err := github.Client.MakeCommit(
				fullScheme,
				stats,
				fmt.Sprintf("Updated to Layer %d", fullScheme.Layer),
			)
			if err != nil {
				return err
			}
			stableDiffs := scheme.GetDiffs(
				environment.LocalStorage.StableLayer,
				fullScheme,
			)
			url, err := telegraph.Client.CreatePage(
				fmt.Sprintf("Layer %d Preview", fullScheme.Layer),
				environment.FormatVar(
					"changelogs",
					map[string]any{
						"differences": differences,
						"stats":       stats,
						"commit_urls": commitUrls,
						"banner_url":  environment.LocalStorage.BannerURL,
						"main_scheme": scheme.ToString(stableDiffs.MainApi, fullScheme.Layer, false),
						"e2e_scheme":  scheme.ToString(stableDiffs.E2EApi, fullScheme.Layer, false),
					},
				),
			)
			if err != nil {
				return err
			}
			if err = bot.Client.UpdateStatus(
				environment.FormatVar(
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
		} else {
			if err = bot.Client.UpdateStatus("", false, false); err != nil {
				return err
			}
		}
		if !slices.Contains(environment.ShellFlags, "debug") {
			if len(environment.LocalStorage.RecentLayers) == 0 {
				environment.LocalStorage.StableLayer = environment.LocalStorage.PreviewLayer
			}
			if !slices.Contains(environment.LocalStorage.RecentLayers, fullScheme.Layer) {
				environment.LocalStorage.RecentLayers = append(environment.LocalStorage.RecentLayers, fullScheme.Layer)
			}
			if len(environment.LocalStorage.RecentLayers) > 1 {
				environment.LocalStorage.RecentLayers = environment.LocalStorage.RecentLayers[1:]
				environment.LocalStorage.StableLayer = environment.LocalStorage.PreviewLayer
			}
			environment.LocalStorage.PreviewLayer = fullScheme
			environment.LocalStorage.Commit()
		}
		return nil
	})
}

package main

import (
	"TLExtractor/android"
	"TLExtractor/appcenter"
	"TLExtractor/appcenter/types"
	"TLExtractor/environment"
	"TLExtractor/github"
	"TLExtractor/java/jadx"
	"TLExtractor/services"
	"TLExtractor/telegram/bot"
	"TLExtractor/telegram/scheme"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"TLExtractor/telegram/telegraph"
	"TLExtractor/tui"
	"TLExtractor/utils"
	"TLExtractor/utils/package_manager"
	"fmt"
	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"slices"
	"time"
)

func main() {
	tui.Run()
	package_manager.CheckPackages()
	services.Run(run)
}

func run() {
	bot.Client.UpdateUptime(true, "")
	if !environment.Debug {
		defer func() {
			if r := recover(); r != nil {
				bot.Client.UpdateUptime(false, "panic")
				gologging.Fatal(r)
			}
		}()
	}
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
			nil,
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
				nil,
			); err != nil {
				bot.Client.UpdateUptime(false, "panic")
				gologging.Fatal(err)
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
			commitInfo, err := github.Client.MakeCommit(
				fullScheme,
				stats,
				fmt.Sprintf("Updated to Layer %d", fullScheme.Layer),
			)
			if err != nil {
				return err
			}
			var stableScheme *schemeTypes.TLFullScheme
			if environment.LocalStorage.StableLayer != nil {
				stableScheme = environment.LocalStorage.StableLayer
			} else {
				stableScheme = environment.LocalStorage.PreviewLayer
			}
			stableDiffs := scheme.GetDiffs(
				stableScheme,
				fullScheme,
			)
			url, err := telegraph.Client.CreatePage(
				fmt.Sprintf("Layer %d Preview", fullScheme.Layer),
				environment.FormatVar(
					"changelogs",
					map[string]any{
						"differences": differences,
						"stats":       stats,
						"commit_urls": commitInfo.FilesLines,
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
				&tgTypes.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgTypes.InlineKeyboardButton{
						{
							{
								Text: "Full Changelog",
								URL:  url,
							},
							{
								Text: "GitHub",
								URL:  commitInfo.SourceURL,
							},
						},
					},
				},
			); err != nil {
				return err
			}
		} else {
			if err = bot.Client.UpdateStatus("", false, false, nil); err != nil {
				return err
			}
		}
		if !environment.Debug {
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

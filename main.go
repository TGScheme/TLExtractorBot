package main

import (
	"TLExtractor/android"
	"TLExtractor/appcenter"
	"TLExtractor/appcenter/types"
	"TLExtractor/consts"
	"TLExtractor/debug_menu"
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
	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/filters"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"slices"
	"time"
)

func main() {
	tui.Run()
	if environment.Debug && !debug_menu.ReadyToTest {
		return
	}
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
	bot.Client.OnMessage(filters.Filter(func(client *gobotapi.Client, update tgTypes.Message) {
		status := environment.IsBuilding()
		if !status {
			environment.SetPatchStatus(true)
		}
		_, _ = client.Invoke(&methods.SendMessage{
			ChatID: update.Chat.ID,
			Text: environment.FormatVar(
				"patch_message",
				map[string]any{
					"is_building": status,
				},
			),
			ParseMode: "HTML",
		})
	}, filters.And(filters.Command("patch", consts.SupportedBotAliases...), filters.ChatID(environment.LocalStorage.LogChatID))))
	appcenter.Listen(func(update types.UpdateInfo) error {
		if err := bot.Client.UpdateStatus(
			environment.FormatVar(
				"message",
				map[string]any{
					"update":   update,
					"progress": 0,
					"is_patch": environment.IsPatch(),
				},
			),
			false,
			false,
			nil,
		); err != nil {
			return err
		}
		startTime := time.Now()
		if update.Source == "android" {
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
							"is_patch": environment.IsPatch(),
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
		}
		elapsedTime := time.Since(startTime)
		var err error
		var fullScheme *schemeTypes.TLFullScheme
		previewLayer := environment.LocalStorage.PreviewLayer.Layer
		if update.Source == "android" {
			fullScheme, err = android.ExtractScheme()
			if err != nil {
				return err
			}
		} else {
			var rawScheme schemeTypes.RawTLScheme
			remoteScheme, err := scheme.GetScheme()
			if err != nil {
				return err
			}
			rawScheme.Layer = remoteScheme.Layer
			rawScheme.Methods = remoteScheme.Methods
			rawScheme.Constructors = remoteScheme.Constructors
			rawScheme.IsSync = remoteScheme.Layer == previewLayer
			fullScheme, err = scheme.MergeUpstream(&rawScheme, schemeTypes.TDesktopPatch, func(isE2E bool) (*schemeTypes.TLRemoteScheme, error) {
				var rScheme schemeTypes.TLRemoteScheme
				var methodsTemp []*schemeTypes.TLMethod
				var constructorsTemp []*schemeTypes.TLConstructor
				var layer = environment.LocalStorage.PreviewLayer
				if isE2E {
					methodsTemp = layer.E2EApi.Methods
					constructorsTemp = layer.E2EApi.Constructors
				} else {
					methodsTemp = layer.MainApi.Methods
					constructorsTemp = layer.MainApi.Constructors
				}
				for _, method := range methodsTemp {
					rScheme.Methods = append(rScheme.Methods, &schemeTypes.TLMethod{
						TLBase: method.TLBase.Clone(),
						Method: method.Method,
					})
				}
				for _, constructor := range constructorsTemp {
					rScheme.Constructors = append(rScheme.Constructors, &schemeTypes.TLConstructor{
						TLBase:    constructor.TLBase.Clone(),
						Predicate: constructor.Predicate,
					})
				}
				rScheme.Layer = previewLayer
				return &rScheme, nil
			})
		}
		if differences := scheme.GetDiffs(environment.LocalStorage.PreviewLayer, fullScheme); differences != nil && fullScheme.Layer >= previewLayer {
			stats := scheme.GetStats(differences)
			commitMessage := fmt.Sprintf("Updated to Layer %d", fullScheme.Layer)
			if environment.IsPatch() {
				commitMessage = fmt.Sprintf("Patch %d", fullScheme.Layer)
			}
			commitInfo, err := github.Client.MakeCommit(
				fullScheme,
				stats,
				commitMessage,
			)
			if err != nil {
				return err
			}
			var stableScheme *schemeTypes.TLFullScheme
			if environment.LocalStorage.StableLayer != nil && slices.Contains(environment.LocalStorage.RecentLayers, fullScheme.Layer) {
				stableScheme = environment.LocalStorage.StableLayer
			} else {
				stableScheme = environment.LocalStorage.PreviewLayer
			}
			stableDiffs := scheme.GetDiffs(
				stableScheme,
				fullScheme,
			)
			pageTitle := fmt.Sprintf("Layer %d", fullScheme.Layer)
			if !fullScheme.IsSync {
				pageTitle = fmt.Sprintf("%s Preview", pageTitle)
			}
			url, err := telegraph.Client.CreatePage(
				pageTitle,
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
						"update":   update,
						"time":     utils.FormatDuration(elapsedTime),
						"layer":    fullScheme.Layer,
						"stats":    stats,
						"is_patch": environment.IsPatch(),
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
		} else {
			if err = bot.Client.UpdateStatus("", false, false, nil); err != nil {
				return err
			}
		}
		return nil
	})
}

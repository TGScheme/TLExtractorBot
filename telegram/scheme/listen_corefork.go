package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/telegram/bot"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
	"fmt"
	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"github.com/anaskhan96/soup"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func ListenCoreFork() {
	Client = &context{}
	chanWait := make(chan bool)
	go func() {
		var isInitialized bool
		for {
			res, err := http.ExecuteRequest(
				fmt.Sprintf("%s/schema", consts.MainReleasedTL),
			)
			if err != nil {
				gologging.Fatal(err)
			}
			var versionsAvailable []int
			if rgx := regexp.MustCompile(`<li><a href="\?layer=([0-9]+)">`).FindAllStringSubmatch(res.String(), -1); len(rgx) > 0 {
				for _, v := range rgx {
					parsedLayer, _ := strconv.Atoi(v[1])
					versionsAvailable = append(versionsAvailable, parsedLayer)
				}
			}
			versionsAvailable = versionsAvailable[1:]
			startLayer := versionsAvailable[0]
			if len(versionsAvailable) == 0 {
				gologging.Fatal("Failed to get the latest version of the TL scheme")
			}
			latestVersion := versionsAvailable[len(versionsAvailable)-1]
			var forceNoUpdate bool
			if environment.LocalStorage.ReleasedLayers == nil {
				environment.LocalStorage.ReleasedLayers = make(map[int]types.ReleasedLayer)
				forceNoUpdate = true
			} else {
				layers := slices.Collect(maps.Keys(environment.LocalStorage.ReleasedLayers))
				slices.Sort(layers)
				startLayer = layers[len(layers)-1]
			}
			if startLayer < latestVersion {
				for _, layer := range versionsAvailable {
					if layer < startLayer {
						continue
					}
					tlRes, err := http.ExecuteRequest(
						fmt.Sprintf("%s/schema/json", consts.MainReleasedTL),
						http.Cookies(map[string]string{
							"stel_dev_layer": strconv.Itoa(layer),
						}),
					)
					if err != nil {
						gologging.Fatal(err)
					}
					var releasedLayer types.ReleasedLayer
					err = json.Unmarshal(tlRes.Body, &releasedLayer)
					if err != nil {
						gologging.Fatal(err)
					}
					environment.LocalStorage.ReleasedLayers[layer] = releasedLayer
				}
			}
			environment.LocalStorage.Commit()
			Client.syncDep.Lock()
			Client.removedConstructors = make([]string, 0)
			checkRemovedConstructors := func(old, new []types.ReleasedConstructor, layer int) {
				var oldConstructors, newConstructors []string
				for _, v := range old {
					oldConstructors = append(oldConstructors, ParseConstructor(v.ID))
				}
				for _, v := range new {
					newConstructors = append(newConstructors, ParseConstructor(v.ID))
				}
				for _, v := range oldConstructors {
					if !slices.Contains(newConstructors, v) {
						Client.removedConstructors = append(Client.removedConstructors, v)
					}
				}
				for i, v := range Client.removedConstructors {
					if slices.Contains(newConstructors, v) {
						Client.removedConstructors = append(Client.removedConstructors[:i], Client.removedConstructors[i+1:]...)
					}
				}
			}
			layers := slices.Collect(maps.Keys(environment.LocalStorage.ReleasedLayers))
			slices.Sort(layers)
			for i := 1; i < len(layers); i++ {
				previousLayer := environment.LocalStorage.ReleasedLayers[layers[i-1]]
				currentLayer := environment.LocalStorage.ReleasedLayers[layers[i]]
				checkRemovedConstructors(previousLayer.Constructors, currentLayer.Constructors, i)
				checkRemovedConstructors(previousLayer.Methods, currentLayer.Methods, i)
			}
			Client.syncDep.Unlock()

			if environment.LocalStorage.LastCoreForkLayer != latestVersion {
				environment.LocalStorage.LastCoreForkLayer = latestVersion
				environment.LocalStorage.Commit()
				if !forceNoUpdate {
					changelogPage := fmt.Sprintf("%s/api/layers", consts.MainReleasedTL)
					res, err = http.ExecuteRequest(changelogPage)
					if err != nil {
						gologging.Fatal(err)
					}
					doc := soup.HTMLParse(res.String())
					devRules := doc.Find("div", "id", "dev_page_content")
					var descriptionText string
					for _, x := range devRules.Children() {
						if x.NodeValue == "h3" && strings.Contains(x.FullText(), strconv.Itoa(latestVersion)) {
							for y := x.Pointer.NextSibling; y != nil && y.Data != "h3"; y = y.NextSibling {
								if y.Data == "ul" {
									rootNode := soup.Root{
										Pointer:   y,
										NodeValue: y.Data,
									}
									descriptionText = rootNode.HTML()
									descriptionText = strings.ReplaceAll(descriptionText, "<li>", "• ")
									descriptionText = strings.ReplaceAll(descriptionText, "</li>", "")
									descriptionText = strings.ReplaceAll(descriptionText, "<ul>", "")
									descriptionText = strings.ReplaceAll(descriptionText, "</ul>", "")
									descriptionText = strings.TrimSpace(descriptionText)
									descriptionText = strings.ReplaceAll(descriptionText, "href=\"/", fmt.Sprintf("href=\"%s/", consts.MainReleasedTL))
									break
								}
							}
							break
						}
					}
					if len(descriptionText) == 0 {
						descriptionText = "• No changelog provided by Telegram MTProto developers."
					}
					err := bot.Client.DirectMessage(
						environment.FormatVar(
							"corefork_update",
							map[string]any{
								"layer":       latestVersion,
								"description": descriptionText,
							},
						),
						&tgTypes.InlineKeyboardMarkup{
							InlineKeyboard: [][]tgTypes.InlineKeyboardButton{
								{
									{
										Text: "Full Changelog",
										URL:  fmt.Sprintf("%s/#layer-%d", changelogPage, latestVersion),
									},
									{
										Text: "Schema",
										URL:  fmt.Sprintf("%s/schema?layer=%d", consts.MainReleasedTL, latestVersion),
									},
								},
							},
						},
					)
					if err != nil {
						gologging.Fatal(err)
					}
				}
			}
			if !isInitialized {
				isInitialized = true
				chanWait <- true
			}
			time.Sleep(1 * time.Second)
		}
	}()
	<-chanWait
}

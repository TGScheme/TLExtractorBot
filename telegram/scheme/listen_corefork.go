package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/telegram/bot"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"github.com/anaskhan96/soup"
)

func init() {
	Client = &context{}
}

func ListenCoreFork() {
	chanWait := make(chan bool)
	go func() {
		var isInitialized bool
		for {
			latestVersion, forceNoUpdate, err := Client.refreshReleasedLayers()
			if err != nil {
				continue
			}
			Client.syncDep.Lock()
			Client.computeRemovedConstructors()
			Client.syncDep.Unlock()

			if latestVersion > environment.LocalStorage.LastCoreForkLayer {
				environment.LocalStorage.LastCoreForkLayer = latestVersion
				environment.LocalStorage.Commit()
				if !forceNoUpdate {
					changelogPage := fmt.Sprintf("%s/api/layers", consts.MainReleasedTL)
					res, err := http.ExecuteRequest(changelogPage)
					if err != nil {
						gologging.Fatal(err)
					}
					doc := soup.HTMLParse(res.String())
					devRules := doc.Find("div", "id", "dev_page_content")
					var descriptionText string
					for _, x := range devRules.Children() {
						if x.NodeValue == "h3" && strings.Contains(x.FullText(), strconv.Itoa(latestVersion)) {
							for y := x.Pointer.NextSibling; y != nil && y.Data != "h3" && y.Data != "h5"; y = y.NextSibling {
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
					err = bot.Client.DirectMessage(
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
						continue
					}
				}
			}
			if !isInitialized {
				isInitialized = true
				chanWait <- true
			}
			time.Sleep(30 * time.Second)
		}
	}()
	<-chanWait
}

func (ctx *context) refreshReleasedLayers() (latestVersion int, forceNoUpdate bool, err error) {
	res, err := http.ExecuteRequest(fmt.Sprintf("%s/schema", consts.MainReleasedTL))
	if err != nil {
		return 0, false, err
	}
	var versionsAvailable []int
	if rgx := regexp.MustCompile(`<li><a href="\?layer=([0-9]+)">`).FindAllStringSubmatch(res.String(), -1); len(rgx) > 0 {
		for _, v := range rgx {
			parsedLayer, _ := strconv.Atoi(v[1])
			versionsAvailable = append(versionsAvailable, parsedLayer)
		}
	}
	if len(versionsAvailable) == 0 {
		return 0, false, errors.New("failed to get the latest version of the TL scheme")
	}
	versionsAvailable = versionsAvailable[1:]
	startLayer := versionsAvailable[0]
	if environment.LocalStorage.ReleasedLayers == nil {
		environment.LocalStorage.ReleasedLayers = make(map[int]types.ReleasedLayer)
		forceNoUpdate = true
	} else {
		layers := slices.Collect(maps.Keys(environment.LocalStorage.ReleasedLayers))
		slices.Sort(layers)
		startLayer = layers[len(layers)-1]
	}
	latestVersion = versionsAvailable[len(versionsAvailable)-1]
	if startLayer < latestVersion {
		for _, layer := range versionsAvailable {
			if layer < startLayer {
				continue
			}
			tlRes, reqErr := http.ExecuteRequest(
				fmt.Sprintf("%s/schema/json", consts.MainReleasedTL),
				http.Cookies(map[string]string{
					"stel_dev_layer": strconv.Itoa(layer),
				}),
			)
			if reqErr != nil {
				return 0, false, reqErr
			}
			var releasedLayer types.ReleasedLayer
			if reqErr = json.Unmarshal(tlRes.Body, &releasedLayer); reqErr != nil {
				return 0, false, reqErr
			}
			environment.LocalStorage.ReleasedLayers[layer] = releasedLayer
		}
	}
	environment.LocalStorage.Commit()
	return latestVersion, forceNoUpdate, nil
}

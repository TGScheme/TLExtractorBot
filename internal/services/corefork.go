package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/banner"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
	"github.com/anaskhan96/soup"
)

var listItemRgx = regexp.MustCompile(`(?s)<li>(.*?)</li>`)

func (s *Service) pollCoreFork() {
	latest, firstRun, err := s.scheme.RefreshReleasedLayers()
	if err != nil {
		gologging.Error(err)
		return
	}
	settings, err := s.db.SettingsStore.GetSettings()
	if err != nil {
		gologging.Error(err)
		return
	}
	if int64(latest) <= settings.LastCoreforkLayer {
		return
	}
	if err = s.db.SettingsStore.SetLastCoreForkLayer(int64(latest)); err != nil {
		gologging.Error(err)
		return
	}
	if firstRun {
		return
	}

	changelogPage := fmt.Sprintf("%s/api/layers", consts.MainReleasedTL)
	changelog := fetchChangelog(changelogPage, latest)
	keyboard := &tgTypes.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgTypes.InlineKeyboardButton{{
			{Text: "Full Changelog", URL: fmt.Sprintf("%s/#layer-%d", changelogPage, latest)},
			{Text: "Schema", URL: fmt.Sprintf("%s/schema?layer=%d", consts.MainReleasedTL, latest)},
		}},
	}
	total := strings.Count(changelog, "<li")
	args := map[string]any{
		"layer":       latest,
		"description": richBullets(changelog),
		"total":       total,
	}
	image := coreforkBanner(latest, total)
	if image != nil {
		args["banner_url"] = bot.BannerSource
	}
	if err = s.bot.DirectRich(assets.Render("corefork_update", args), keyboard, image); err != nil {
		gologging.Error(err)
	}
}

func coreforkBanner(layer, total int) []byte {
	image, err := banner.Render(banner.Input{
		Layer:        layer,
		Title:        "Released on Corefork",
		Source:       "corefork.telegram.org",
		ChangesLabel: "CHANGELOG",
		Highlight:    fmt.Sprintf("%d", total),
		Changes:      " entries published by Telegram",
		IsStable:     true,
	})
	if err != nil {
		gologging.Error("banner: unable to render the corefork banner:", err)
		return nil
	}
	return image
}

func richBullets(list string) string {
	items := listItemRgx.FindAllStringSubmatch(list, -1)
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, "• "+strings.Join(strings.Fields(item[1]), " "))
	}
	return strings.Join(rows, "<br>")
}

func fetchChangelog(page string, layer int) string {
	res, err := http.ExecuteRequest(page)
	if err != nil {
		return ""
	}
	content := soup.HTMLParse(res.String()).Find("div", "id", "dev_page_content")
	for _, node := range content.Children() {
		if node.NodeValue != "h3" || !strings.Contains(node.FullText(), strconv.Itoa(layer)) {
			continue
		}
		for sibling := node.Pointer.NextSibling; sibling != nil && sibling.Data != "h3" && sibling.Data != "h5"; sibling = sibling.NextSibling {
			if sibling.Data != "ul" {
				continue
			}
			text := soup.Root{Pointer: sibling, NodeValue: sibling.Data}.HTML()
			return strings.ReplaceAll(
				strings.TrimSpace(text),
				"href=\"/", fmt.Sprintf("href=\"%s/", consts.MainReleasedTL),
			)
		}
		break
	}
	return ""
}

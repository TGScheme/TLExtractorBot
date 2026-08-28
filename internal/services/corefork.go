package services

import (
	"fmt"
	"strconv"
	"strings"

	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/anaskhan96/soup"
)

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
	if err = s.bot.DirectRich(assets.Render("corefork_update", map[string]any{
		"layer":       latest,
		"description": changelog,
		"total":       strings.Count(changelog, "<li"),
	}), keyboard); err == nil {
		return
	}
	gologging.Error("telegram: unable to send the corefork rich message, falling back to the plain one:", err)
	if err = s.bot.DirectMessage(assets.Render("corefork_update_plain", map[string]any{
		"layer":       latest,
		"description": bulletize(changelog),
	}), keyboard); err != nil {
		gologging.Error(err)
	}
}

func bulletize(list string) string {
	for _, replacement := range [][2]string{
		{"<li>", "• "}, {"</li>", ""}, {"<ul>", ""}, {"</ul>", ""},
	} {
		list = strings.ReplaceAll(list, replacement[0], replacement[1])
	}
	return strings.TrimSpace(list)
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

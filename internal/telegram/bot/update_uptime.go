package bot

import (
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/utils"
	"time"

	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/Laky-64/gologging"
)

func (ctx *Client) UpdateUptime(online bool, exitReason string) {
	_, err := ctx.client.Invoke(
		&methods.SendMessage{
			ChatID: ctx.logChatID,
			Text: assets.Render(
				"uptime",
				map[string]any{
					"online":      online,
					"uptime":      utils.FormatDuration(time.Since(ctx.startedAt)),
					"exit_reason": exitReason,
				},
			),
			ParseMode: "html",
		},
	)
	if err != nil {
		gologging.Fatal(err)
	}
}

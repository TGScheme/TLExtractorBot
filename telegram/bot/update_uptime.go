package bot

import (
	"TLExtractor/environment"
	"TLExtractor/utils"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/Laky-64/gologging"
	"time"
)

func (ctx *context) UpdateUptime(online bool, exitReason string) {
	_, err := ctx.client.Invoke(
		&methods.SendMessage{
			ChatID: environment.LocalStorage.LogChatID,
			Text: environment.FormatVar(
				"uptime",
				map[string]any{
					"online":      online,
					"uptime":      utils.FormatDuration(time.Since(environment.StartTime)),
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

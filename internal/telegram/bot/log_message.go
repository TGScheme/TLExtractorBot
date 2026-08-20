package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *Client) LogMessage(text string) error {
	_, err := ctx.client.Invoke(
		&methods.SendMessage{
			ChatID:    ctx.logChatID,
			Text:      text,
			ParseMode: "HTML",
			LinkPreviewOptions: &types.LinkPreviewOptions{
				IsDisabled: true,
			},
		},
	)
	return err
}

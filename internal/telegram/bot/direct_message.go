package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *Client) DirectMessage(text string, keyboard *types.InlineKeyboardMarkup) error {
	_, err := ctx.client.Invoke(
		&methods.SendMessage{
			ChatID:    ctx.channelID,
			Text:      text,
			ParseMode: "HTML",
			LinkPreviewOptions: &types.LinkPreviewOptions{
				IsDisabled: true,
			},
			ReplyMarkup: keyboard,
		},
	)
	return err
}

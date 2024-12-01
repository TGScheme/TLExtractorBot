package bot

import (
	"TLExtractor/environment"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *context) DirectMessage(text string, keyboard *types.InlineKeyboardMarkup) error {
	_, err := ctx.client.Invoke(
		&methods.SendMessage{
			ChatID:    environment.LocalStorage.ChannelID,
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

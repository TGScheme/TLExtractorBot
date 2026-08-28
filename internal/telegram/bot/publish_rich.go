package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *Client) PublishRich(html string, withNotification bool, keyboard *types.InlineKeyboardMarkup) error {
	if ctx.statusMessageID != 0 {
		if _, err := ctx.client.Invoke(&methods.DeleteMessage{
			ChatID:    ctx.channelID,
			MessageID: ctx.statusMessageID,
		}); err != nil {
			return err
		}
		ctx.statusMessageID = 0
	}
	ctx.statusText = ""
	_, err := ctx.client.Invoke(&methods.SendRichMessage{
		ChatID:              ctx.channelID,
		RichMessage:         types.InputRichMessage{Html: html},
		DisableNotification: !withNotification,
		ReplyMarkup:         keyboard,
	})
	return err
}

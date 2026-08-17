package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *Client) UpdateStatus(text string, withNotification, isFinal bool, keyboard *types.InlineKeyboardMarkup) error {
	if len(text) == 0 {
		_, err := ctx.client.Invoke(
			&methods.DeleteMessage{
				ChatID:    ctx.channelID,
				MessageID: ctx.statusMessageID,
			},
		)
		if err != nil {
			return err
		}
		ctx.statusMessageID = 0
	} else {
		if ctx.statusMessageID != 0 {
			if isFinal {
				_, err := ctx.client.Invoke(
					&methods.DeleteMessage{
						ChatID:    ctx.channelID,
						MessageID: ctx.statusMessageID,
					},
				)
				if err != nil {
					return err
				}
				ctx.statusMessageID = 0
			} else {
				_, err := ctx.client.Invoke(
					&methods.EditMessageText{
						ChatID:    ctx.channelID,
						MessageID: ctx.statusMessageID,
						Text:      text,
						ParseMode: "HTML",
					},
				)
				if err == nil {
					return nil
				}
			}
		}
		res, err := ctx.client.Invoke(
			&methods.SendMessage{
				ChatID:              ctx.channelID,
				Text:                text,
				DisableNotification: !withNotification,
				ParseMode:           "HTML",
				LinkPreviewOptions: &types.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyMarkup: keyboard,
			},
		)
		if err != nil {
			return err
		}
		if !isFinal {
			ctx.statusMessageID = res.Result.(types.Message).MessageID
		}
	}
	return nil
}

package bot

import (
	"TLExtractor/environment"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *context) UpdateStatus(text string, withNotification, isFinal bool, keyboard *types.InlineKeyboardMarkup) error {
	if len(text) == 0 {
		_, err := ctx.client.Invoke(
			&methods.DeleteMessage{
				ChatID:    environment.LocalStorage.ChannelID,
				MessageID: environment.LocalStorage.MessageId,
			},
		)
		if err != nil {
			return err
		}
		environment.LocalStorage.MessageId = 0
	} else {
		if environment.LocalStorage.MessageId != 0 {
			if isFinal {
				_, err := ctx.client.Invoke(
					&methods.DeleteMessage{
						ChatID:    environment.LocalStorage.ChannelID,
						MessageID: environment.LocalStorage.MessageId,
					},
				)
				if err != nil {
					return err
				}
				environment.LocalStorage.MessageId = 0
			} else {
				_, err := ctx.client.Invoke(
					&methods.EditMessageText{
						ChatID:    environment.LocalStorage.ChannelID,
						MessageID: environment.LocalStorage.MessageId,
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
				ChatID:              environment.LocalStorage.ChannelID,
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
			environment.LocalStorage.MessageId = res.Result.(types.Message).MessageID
		}
	}
	environment.LocalStorage.Commit()
	return nil
}

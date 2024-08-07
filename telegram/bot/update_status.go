package bot

import (
	"TLExtractor/utils"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *Context) UpdateStatus(text string, withNotification, isFinal bool) error {
	if utils.LocalStorage.MessageId != 0 {
		if isFinal {
			_, err := ctx.client.Invoke(
				&methods.DeleteMessage{
					ChatID:    utils.LocalStorage.ChannelID,
					MessageID: utils.LocalStorage.MessageId,
				},
			)
			if err != nil {
				return err
			}
			utils.LocalStorage.MessageId = 0
		} else {
			_, err := ctx.client.Invoke(
				&methods.EditMessageText{
					ChatID:    utils.LocalStorage.ChannelID,
					MessageID: utils.LocalStorage.MessageId,
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
			ChatID:              utils.LocalStorage.ChannelID,
			Text:                text,
			DisableNotification: !withNotification,
			ParseMode:           "HTML",
			LinkPreviewOptions: &types.LinkPreviewOptions{
				IsDisabled: true,
			},
		},
	)
	if err != nil {
		return err
	}
	if !isFinal {
		utils.LocalStorage.MessageId = res.Result.(types.Message).MessageID
	}
	return utils.LocalStorage.Commit()
}

package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *Client) PublishRich(html string, withNotification bool, keyboard *types.InlineKeyboardMarkup, banner []byte) error {
	ctx.discardStatus()
	ctx.statusMutex.Lock()
	defer ctx.statusMutex.Unlock()
	if err := ctx.dropStatus(); err != nil {
		return err
	}
	return ctx.sendRich(&methods.SendRichMessage{
		ChatID:              ctx.channelID,
		RichMessage:         types.InputRichMessage{Html: html},
		DisableNotification: !withNotification,
		ReplyMarkup:         keyboard,
	}, banner)
}

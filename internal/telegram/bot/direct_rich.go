package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *Client) DirectRich(html string, keyboard *types.InlineKeyboardMarkup, banner []byte) error {
	return ctx.sendRich(&methods.SendRichMessage{
		ChatID:      ctx.channelID,
		RichMessage: types.InputRichMessage{Html: html},
		ReplyMarkup: keyboard,
	}, banner)
}

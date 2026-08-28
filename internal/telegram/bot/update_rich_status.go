package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
)

func (ctx *Client) UpdateRichStatus(html, plain string) error {
	if ctx.statusMessageID != 0 {
		if html == ctx.statusText {
			return nil
		}
		if _, err := ctx.client.Invoke(&methods.EditMessageText{
			ChatID:      ctx.channelID,
			MessageID:   ctx.statusMessageID,
			RichMessage: &types.InputRichMessage{Html: html},
		}); err == nil {
			ctx.statusText = html
			return nil
		}
		ctx.dropStatus()
	}
	res, err := ctx.client.Invoke(&methods.SendRichMessage{
		ChatID:              ctx.channelID,
		RichMessage:         types.InputRichMessage{Html: html},
		DisableNotification: true,
	})
	if err != nil {
		gologging.Error("telegram: unable to send the rich status, falling back to the plain one:", err)
		return ctx.UpdateStatus(plain, false, false, nil)
	}
	ctx.statusMessageID = res.Result.(types.Message).MessageID
	ctx.statusText = html
	return nil
}

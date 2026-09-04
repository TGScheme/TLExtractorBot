package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
)

func (ctx *Client) UpdateRichStatus(html string) error {
	ctx.statusMutex.Lock()
	defer ctx.statusMutex.Unlock()
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
		if err := ctx.dropStatus(); err != nil {
			gologging.Error("telegram: unable to delete the stale status message:", err)
		}
	}
	res, err := ctx.client.Invoke(&methods.SendRichMessage{
		ChatID:              ctx.channelID,
		RichMessage:         types.InputRichMessage{Html: html},
		DisableNotification: true,
	})
	if err != nil {
		return err
	}
	ctx.statusMessageID = res.Result.(types.Message).MessageID
	ctx.statusText = html
	return nil
}

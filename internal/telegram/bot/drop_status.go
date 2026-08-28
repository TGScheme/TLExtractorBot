package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/Laky-64/gologging"
)

func (ctx *Client) dropStatus() {
	if ctx.statusMessageID == 0 {
		return
	}
	if _, err := ctx.client.Invoke(&methods.DeleteMessage{
		ChatID:    ctx.channelID,
		MessageID: ctx.statusMessageID,
	}); err != nil {
		gologging.Error("telegram: unable to delete the stale status message:", err)
	}
	ctx.statusMessageID = 0
	ctx.statusText = ""
}

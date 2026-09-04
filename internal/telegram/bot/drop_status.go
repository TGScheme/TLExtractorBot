package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
)

func (ctx *Client) DropStatus() error {
	ctx.statusMutex.Lock()
	defer ctx.statusMutex.Unlock()
	return ctx.dropStatus()
}

func (ctx *Client) dropStatus() error {
	if ctx.statusMessageID == 0 {
		return nil
	}
	if _, err := ctx.client.Invoke(&methods.DeleteMessage{
		ChatID:    ctx.channelID,
		MessageID: ctx.statusMessageID,
	}); err != nil {
		return err
	}
	ctx.statusMessageID = 0
	ctx.statusText = ""
	return nil
}

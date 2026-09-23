package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/Laky-64/gologging"
)

func (ctx *Client) UpdateRichStatus(html string) {
	ctx.statusQueue.Lock()
	ctx.statusPending, ctx.statusQueued = html, true
	ctx.statusQueue.Unlock()
	select {
	case ctx.statusWake <- struct{}{}:
	default:
	}
}

func (ctx *Client) runStatus() {
	for range ctx.statusWake {
		ctx.statusQueue.Lock()
		html, queued, generation := ctx.statusPending, ctx.statusQueued, ctx.statusGeneration
		ctx.statusQueued = false
		ctx.statusQueue.Unlock()
		if !queued {
			continue
		}
		ctx.statusMutex.Lock()
		ctx.statusQueue.Lock()
		current := ctx.statusGeneration == generation
		ctx.statusQueue.Unlock()
		if current {
			if err := ctx.applyStatus(html); err != nil {
				gologging.Error("telegram: unable to update the status message:", err)
			}
		}
		ctx.statusMutex.Unlock()
	}
}

func (ctx *Client) discardStatus() {
	ctx.statusQueue.Lock()
	ctx.statusQueued = false
	ctx.statusGeneration++
	ctx.statusQueue.Unlock()
}

func (ctx *Client) applyStatus(html string) error {
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

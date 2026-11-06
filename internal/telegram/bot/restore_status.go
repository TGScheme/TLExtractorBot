package bot

import "github.com/Laky-64/gologging"

func (ctx *Client) RestoreStatus(messageID int64, persist func(int64) error) {
	ctx.statusMutex.Lock()
	defer ctx.statusMutex.Unlock()
	ctx.statusMessageID = messageID
	ctx.statusText = ""
	ctx.statusPersist = persist
}

func (ctx *Client) setStatusMessageID(messageID int64) {
	if ctx.statusMessageID == messageID {
		return
	}
	ctx.statusMessageID = messageID
	if ctx.statusPersist == nil {
		return
	}
	if err := ctx.statusPersist(messageID); err != nil {
		gologging.Error("telegram: unable to save the status message id:", err)
	}
}

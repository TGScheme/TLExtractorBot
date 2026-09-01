package bot

import (
	"fmt"

	"github.com/gotd/td/tg"
)

func (ctx *Client) mtProtoClient() (*tg.Client, error) {
	ctx.mtProtoMutex.RLock()
	defer ctx.mtProtoMutex.RUnlock()
	if ctx.mtProtoAPI == nil {
		return nil, fmt.Errorf("mtproto session not connected")
	}
	return ctx.mtProtoAPI, nil
}

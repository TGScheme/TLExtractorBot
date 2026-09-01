package bot

import (
	"fmt"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
)

func (ctx *Client) superviseMTProto(cfg *config.Config, ready chan<- error) {
	first := true
	for ctx.mtProtoCtx.Err() == nil {
		err := ctx.connectMTProto(cfg, func() {
			if first {
				first = false
				ready <- nil
			}
		})
		if ctx.mtProtoCtx.Err() != nil {
			return
		}
		if first {
			first = false
			ready <- err
			return
		}
		ctx.mtProtoMutex.Lock()
		ctx.mtProtoAPI = nil
		clear(ctx.channels)
		ctx.mtProtoMutex.Unlock()
		gologging.Error(fmt.Errorf("mtproto disconnected: %w", err))
		select {
		case <-ctx.mtProtoCtx.Done():
			return
		case <-time.After(consts.MTProtoReconnectDelay):
		}
	}
}

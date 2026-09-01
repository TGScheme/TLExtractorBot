package bot

import (
	"context"
	"path"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
)

func (ctx *Client) connectMTProto(cfg *config.Config, onReady func()) error {
	client := telegram.NewClient(cfg.APIID, cfg.APIHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: path.Join(cfg.WorkDir, consts.MTProtoSessionFile),
		},
	})
	return client.Run(ctx.mtProtoCtx, func(runCtx context.Context) error {
		if _, err := client.Auth().Bot(runCtx, cfg.BotToken); err != nil {
			return err
		}
		ctx.mtProtoMutex.Lock()
		ctx.mtProtoAPI = client.API()
		ctx.mtProtoMutex.Unlock()
		gologging.Info("mtproto: session ready")
		onReady()
		<-runCtx.Done()
		return runCtx.Err()
	})
}

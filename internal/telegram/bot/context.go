package bot

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/logger"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/gotd/td/tg"
)

type Client struct {
	client    *gobotapi.PollingClient
	channelID int64
	logChatID int64
	startedAt time.Time

	statusMessageID int64
	statusText      string

	mtProtoMutex  sync.RWMutex
	mtProtoCtx    context.Context
	mtProtoCancel context.CancelFunc
	mtProtoAPI    *tg.Client
	channels      map[string]*tg.InputChannel
}

type ChannelPost struct {
	ID       int
	Text     string
	Document *tg.Document
}

func New(cfg *config.Config) (*Client, error) {
	client := gobotapi.NewClient(cfg.BotToken)
	client.NoNotice = true
	client.LoggingLevel = logger.Silent
	if err := client.Start(); err != nil {
		return nil, err
	}
	if _, err := client.Invoke(&methods.GetMe{}); err != nil {
		return nil, errors.New("invalid bot token")
	}
	mtProtoCtx, mtProtoCancel := context.WithCancel(context.Background())
	ctx := &Client{
		client:        client,
		channelID:     cfg.ChannelID,
		logChatID:     cfg.LogChatID,
		startedAt:     time.Now(),
		mtProtoCtx:    mtProtoCtx,
		mtProtoCancel: mtProtoCancel,
		channels:      make(map[string]*tg.InputChannel),
	}
	ready := make(chan error, 1)
	go ctx.superviseMTProto(cfg, ready)
	if err := <-ready; err != nil {
		mtProtoCancel()
		return nil, err
	}
	return ctx, nil
}

func (ctx *Client) Close() {
	ctx.mtProtoCancel()
}

package bot

import (
	"errors"
	"time"

	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/logger"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/TGScheme/TLExtractorBot/internal/config"
)

type Client struct {
	client    *gobotapi.PollingClient
	channelID int64
	logChatID int64
	startedAt time.Time

	statusMessageID int64
	statusText      string
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
	return &Client{
		client:    client,
		channelID: cfg.ChannelID,
		logChatID: cfg.LogChatID,
		startedAt: time.Now(),
	}, nil
}

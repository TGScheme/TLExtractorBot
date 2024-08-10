package services

import (
	"TLExtractor/telegram/bot"
	"github.com/kardianos/service"
)

func (c *context) Stop(_ service.Service) error {
	bot.Client.UpdateUptime(false, "service_stopped")
	return nil
}

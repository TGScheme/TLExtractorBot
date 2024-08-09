package services

import (
	"TLExtractor/telegram/bot"
	"github.com/kardianos/service"
)

func (c *context) Stop(s service.Service) error {
	bot.Client.UpdateUptime(false, "service_stopped")
	return nil
}

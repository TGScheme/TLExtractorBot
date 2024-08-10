package services

import (
	"github.com/kardianos/service"
)

func (c *context) Start(_ service.Service) error {
	go c.funcRun()
	return nil
}

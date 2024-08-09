package bot

import (
	"github.com/GoBotApiOfficial/gobotapi"
)

var Client *context

type context struct {
	client *gobotapi.PollingClient
}

package bot

import (
	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *context) OnMessage(handler func(client *gobotapi.Client, update types.Message)) {
	ctx.client.OnMessage(handler)
}

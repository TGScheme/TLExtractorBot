package bot

import (
	"TLExtractor/environment"
	"TLExtractor/logging"
	"fmt"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

func (ctx *context) UpdateUptime(online bool) {
	if len(environment.LocalStorage.BotName) == 0 {
		invoke, err := ctx.client.Invoke(
			&methods.GetMyName{},
		)
		if err != nil {
			logging.Fatal(err)
		}
		environment.LocalStorage.BotName = invoke.Result.(types.BotName).Name
		environment.LocalStorage.Commit()
	}
	var status string
	if online {
		status = "ONLINE"
	} else {
		status = "OFFLINE"
	}
	_, err := ctx.client.Invoke(
		&methods.SetMyName{
			Name: fmt.Sprintf("%s [%s]", environment.LocalStorage.BotName, status),
		},
	)
	if err != nil {
		logging.Fatal(err)
	}
}

package bot

import (
	"TLExtractor/consts"
	"TLExtractor/io"
	"TLExtractor/utils"
	"errors"
	"fmt"
	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/logger"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
	"strconv"
	"strings"
)

func NewClient() (*Context, error) {
	var ctx Context
	var bot *gobotapi.PollingClient
	for {
		if len(utils.CredentialsStorage.BotToken) == 0 {
			fmt.Print("Enter bot token: ")
			_ = io.Scanln(&utils.CredentialsStorage.BotToken)
		}
		bot = gobotapi.NewClient(utils.CredentialsStorage.BotToken)
		bot.NoUpdates = true
		bot.LoggingLevel = logger.Silent
		_ = bot.Start()
		if _, err := bot.Invoke(&methods.GetMe{}); err != nil {
			utils.CredentialsStorage.BotToken = ""
			utils.CrashLog(consts.InvalidToken, false)
			continue
		}
		break
	}
	if err := utils.CredentialsStorage.Commit(); err != nil {
		return nil, err
	}
	channelID := strconv.Itoa(int(utils.LocalStorage.ChannelID))
	for {
		if utils.LocalStorage.ChannelID == 0 {
			fmt.Print("Enter channel ID or username: ")
			_ = io.Scanln(&channelID)
			if !strings.HasPrefix(channelID, "@") {
				if _, err := strconv.Atoi(channelID); err != nil {
					channelID = "@" + channelID
				}
			}
		}
		result, err := bot.Invoke(
			&methods.GetChat{
				ChatID: channelID,
			},
		)
		if err != nil {
			utils.LocalStorage.ChannelID = 0
			utils.CrashLog(errors.New("channel not found"), false)
			continue
		}
		utils.LocalStorage.ChannelID = result.Result.(types.ChatFullInfo).ID
		break
	}
	if err := utils.LocalStorage.Commit(); err != nil {
		return nil, err
	}
	ctx.client = bot
	return &ctx, nil
}

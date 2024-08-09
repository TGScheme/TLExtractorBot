package bot

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/io"
	"TLExtractor/logging"
	"errors"
	"fmt"
	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/logger"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/kardianos/service"
	"strconv"
	"strings"
)

func init() {
	Client = &context{}
	var bot *gobotapi.PollingClient
	for {
		if len(environment.CredentialsStorage.BotToken) == 0 {
			if !service.Interactive() {
				logging.Fatal("Bot token is required")
			}
			fmt.Print("Enter bot token: ")
			_ = io.Scanln(&environment.CredentialsStorage.BotToken)
		}
		bot = gobotapi.NewClient(environment.CredentialsStorage.BotToken)
		bot.NoUpdates = true
		bot.LoggingLevel = logger.Silent
		_ = bot.Start()
		if _, err := bot.Invoke(&methods.GetMe{}); err != nil {
			environment.CredentialsStorage.BotToken = ""
			logging.Error(consts.InvalidToken)
			continue
		}
		break
	}
	environment.CredentialsStorage.Commit()
	channelID := strconv.Itoa(int(environment.LocalStorage.ChannelID))
	for {
		if environment.LocalStorage.ChannelID == 0 {
			if !service.Interactive() {
				logging.Fatal("Channel ID is required")
			}
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
			environment.LocalStorage.ChannelID = 0
			logging.Error(errors.New("channel not found"))
			continue
		}
		environment.LocalStorage.ChannelID = result.Result.(types.ChatFullInfo).ID
		break
	}
	logChatId := strconv.Itoa(int(environment.LocalStorage.LogChatID))
	for {
		if environment.LocalStorage.LogChatID == 0 {
			if !service.Interactive() {
				logging.Fatal("Log chat ID is required")
			}
			fmt.Print("Enter log chat ID or username: ")
			_ = io.Scanln(&logChatId)
			if !strings.HasPrefix(logChatId, "@") {
				if _, err := strconv.Atoi(logChatId); err != nil {
					logChatId = "@" + logChatId
				}
			}
		} else {
			break
		}
		result, err := bot.Invoke(
			&methods.SendMessage{
				ChatID: logChatId,
				Text:   "Checking log chat ID..",
			},
		)
		if err != nil {
			environment.LocalStorage.ChannelID = 0
			logging.Error(errors.New("chat not found"))
			continue
		}
		_, _ = bot.Invoke(
			&methods.DeleteMessage{
				ChatID:    environment.LocalStorage.LogChatID,
				MessageID: result.Result.(types.Message).MessageID,
			},
		)
		environment.LocalStorage.LogChatID = result.Result.(types.Message).Chat.ID
		break
	}
	environment.LocalStorage.Commit()
	Client.client = bot
}

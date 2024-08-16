package bot

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/tui"
	tuiTypes "TLExtractor/tui/types"
	"errors"
	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/logger"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/charmbracelet/huh"
	"strconv"
	"strings"
)

func init() {
	Client = &context{}
	var bot *gobotapi.PollingClient
	telegramApp := tui.NewMiniApp("telegram")
	telegramApp.SetLoadingMessage("Logging in to Telegram Bot API...")
	telegramApp.SetFields(
		huh.NewInput().
			Title("Bot token").
			Description("Enter your bot token").
			Placeholder("123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11").
			EchoMode(huh.EchoModePassword).
			Validate(tui.Validate("Bot token", tuiTypes.NoCheck)).
			Value(&environment.CredentialsStorage.BotToken),
	)
	telegramApp.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		if len(environment.CredentialsStorage.BotToken) == 0 {
			return errors.New("telegram token is required")
		}
		if checkType == tuiTypes.InitCheck {
			if environment.LocalStorage.ChannelID == 0 || environment.LocalStorage.LogChatID == 0 {
				return errors.New("channel ID and log chat ID are required")
			}
		}
		bot = gobotapi.NewClient(environment.CredentialsStorage.BotToken)
		bot.NoNotice = true
		bot.LoggingLevel = logger.Silent
		_ = bot.Start()
		if _, err := bot.Invoke(&methods.GetMe{}); err != nil {
			environment.CredentialsStorage.BotToken = ""
			return consts.InvalidToken
		}
		Client.client = bot
		environment.CredentialsStorage.Commit()
		return nil
	}, tuiTypes.InitCheck, tuiTypes.SubmitCheck)

	var channel, logChat string
	infoPage := telegramApp.NewAppPage()
	infoPage.SetFields(
		huh.NewInput().
			Title("Channel ID").
			Description("Enter channel ID or username where the bot will send messages").
			Placeholder("@channel").
			Validate(tui.Validate("Channel ID", tuiTypes.NoCheck)).
			Value(&channel),
		huh.NewInput().
			Title("Log chat ID").
			Description("Enter log User ID or channel username where the bot will send logs").
			Placeholder("@channel").
			Validate(tui.Validate("Log chat ID", tuiTypes.NoCheck)).
			Value(&logChat),
	)
	infoPage.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		parsePeer := func(value string) string {
			if !strings.HasPrefix(value, "@") {
				if _, err := strconv.Atoi(value); err != nil {
					return "@" + value
				}
			}
			return value
		}
		channelPeer, logPeer := parsePeer(channel), parsePeer(logChat)
		result, err := bot.Invoke(
			&methods.GetChat{
				ChatID: channelPeer,
			},
		)
		if err != nil {
			return errors.New("channel not found")
		}
		environment.LocalStorage.ChannelID = result.Result.(types.ChatFullInfo).ID
		result, err = bot.Invoke(
			&methods.SendMessage{
				ChatID: logPeer,
				Text:   "Checking log chat ID..",
			},
		)
		if err != nil {
			return errors.New("chat not found")
		}
		_, _ = bot.Invoke(
			&methods.DeleteMessage{
				ChatID:    logPeer,
				MessageID: result.Result.(types.Message).MessageID,
			},
		)
		environment.LocalStorage.LogChatID = result.Result.(types.Message).Chat.ID
		environment.LocalStorage.Commit()
		return nil
	}, tuiTypes.SubmitCheck)
}

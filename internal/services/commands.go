package services

import (
	"fmt"
	"strings"

	"github.com/GoBotApiOfficial/gobotapi"
	"github.com/GoBotApiOfficial/gobotapi/filters"
	"github.com/GoBotApiOfficial/gobotapi/methods"
	tgTypes "github.com/GoBotApiOfficial/gobotapi/types"
	"github.com/TGScheme/TLExtractorBot/internal/assets"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/gemini"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
)

func RegisterCommands(
	cfg *config.Config,
	database *db.DB,
	botClient *bot.Client,
	geminiClient *gemini.Client,
	service *Service,
) {
	botClient.OnMessage(filters.Filter(func(client *gobotapi.Client, update tgTypes.Message) {
		building := service.IsBuilding()
		if !building {
			service.RequestPatch()
		}
		reply(client, update, assets.Render("patch_message", map[string]any{
			"is_building": building,
		}))
	}, filters.And(filters.Command("patch", supportedAliases...), filters.ChatID(cfg.LogChatID))))

	botClient.OnMessage(filters.Filter(func(client *gobotapi.Client, update tgTypes.Message) {
		models, err := geminiClient.Models()
		if err != nil {
			reply(client, update, fmt.Sprintf("Cannot list models: %v", err))
			return
		}
		requested := commandArgument(update.Text)
		if requested == "" {
			var lines []string
			for _, model := range models {
				marker := ""
				if model.ID == geminiClient.Model() {
					marker = " ✅"
				}
				lines = append(lines, fmt.Sprintf("<code>%s</code>%s", model.ID, marker))
			}
			reply(client, update, "Available models:\n"+strings.Join(lines, "\n"))
			return
		}
		for _, model := range models {
			if model.ID != requested {
				continue
			}
			if err = database.SettingsStore.SetLLMModel(model.ID); err != nil {
				reply(client, update, fmt.Sprintf("Cannot save model: %v", err))
				return
			}
			geminiClient.SetModel(model.ID)
			reply(client, update, fmt.Sprintf("Model set to <code>%s</code>", model.ID))
			return
		}
		reply(client, update, fmt.Sprintf("Unknown model <code>%s</code>", requested))
	}, filters.And(filters.Command("model", supportedAliases...), filters.ChatID(cfg.LogChatID))))

	botClient.OnMessage(filters.Filter(func(client *gobotapi.Client, update tgTypes.Message) {
		branch := commandArgument(update.Text)
		if branch == "" {
			settings, err := database.SettingsStore.GetSettings()
			if err != nil {
				reply(client, update, fmt.Sprintf("Cannot read settings: %v", err))
				return
			}
			reply(client, update, fmt.Sprintf("tdesktop branch: <code>%s</code>", settings.TdesktopBranch))
			return
		}
		if err := database.SettingsStore.SetTDesktopBranch(branch); err != nil {
			reply(client, update, fmt.Sprintf("Cannot save branch: %v", err))
			return
		}
		reply(client, update, fmt.Sprintf("tdesktop branch set to <code>%s</code>", branch))
	}, filters.And(filters.Command("branch", supportedAliases...), filters.ChatID(cfg.LogChatID))))

	botClient.OnMessage(filters.Filter(func(client *gobotapi.Client, update tgTypes.Message) {
		data, filename, err := service.Backup()
		if err != nil {
			reply(client, update, fmt.Sprintf("Backup failed: %v", err))
			return
		}
		if err = botClient.SendDocument(update.Chat.ID, filename, data, "<b>Manual backup</b>"); err != nil {
			reply(client, update, fmt.Sprintf("Cannot upload backup: %v", err))
		}
	}, filters.And(filters.Command("backup", supportedAliases...), filters.ChatID(cfg.LogChatID))))

	botClient.OnMessage(filters.Filter(func(client *gobotapi.Client, update tgTypes.Message) {
		data, err := client.DownloadBytes(update.Document.FileID, nil)
		if err != nil {
			reply(client, update, fmt.Sprintf("Cannot download the dump: %v", err))
			return
		}
		if err = service.RestoreBackup(data); err != nil {
			reply(client, update, fmt.Sprintf("Restore failed: %v", err))
			return
		}
		reply(client, update, "<b>✅ Backup restored</b>")
	}, filters.And(filters.ChatID(cfg.LogChatID), isBackupFile())))
}

func isBackupFile() filters.FilterOperand {
	return func(df *filters.DataFilter) bool {
		message, ok := df.RawUpdate.(tgTypes.Message)
		if !ok || message.Document == nil {
			return false
		}
		name := message.Document.FileName
		return strings.HasSuffix(name, ".sql") || strings.HasSuffix(name, ".dump")
	}
}

var supportedAliases = []string{".", "/", "!"}

func commandArgument(text string) string {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func reply(client *gobotapi.Client, update tgTypes.Message, text string) {
	_, _ = client.Invoke(&methods.SendMessage{
		ChatID:    update.Chat.ID,
		Text:      text,
		ParseMode: "HTML",
	})
}

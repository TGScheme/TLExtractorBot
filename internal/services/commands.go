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
	"github.com/TGScheme/TLExtractorBot/internal/llm"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
)

func RegisterCommands(
	cfg *config.Config,
	database *db.DB,
	botClient *bot.Client,
	llmClient *llm.Client,
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
		models, err := llmClient.Models()
		if err != nil {
			reply(client, update, fmt.Sprintf("Cannot list models: %v", err))
			return
		}
		requested := commandArgument(update.Text)
		if requested == "" {
			var sections []string
			for _, kind := range llmClient.Kinds() {
				lines := []string{fmt.Sprintf("<b>%s models</b>", kind)}
				for _, model := range models {
					if model.Kind != kind {
						continue
					}
					marker := ""
					if model.ID == llmClient.Model(kind) {
						marker = " ✅"
					}
					lines = append(lines, fmt.Sprintf("<code>%s</code>%s", model.ID, marker))
				}
				sections = append(sections, strings.Join(lines, "\n"))
			}
			reply(client, update, strings.Join(sections, "\n\n"))
			return
		}
		for _, model := range models {
			if model.ID != requested {
				continue
			}
			save := database.SettingsStore.SetLLMModel
			if model.Kind == llm.KindOpenAI {
				save = database.SettingsStore.SetOpenAIModel
			}
			if err = save(model.ID); err != nil {
				reply(client, update, fmt.Sprintf("Cannot save model: %v", err))
				return
			}
			llmClient.SetModel(model.Kind, model.ID)
			reply(client, update, fmt.Sprintf("%s model set to <code>%s</code>", model.Kind, model.ID))
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
		go service.ExtractUploadedApk(int(update.MessageID), isPatchCaption(update.Caption), func(text string) {
			reply(client, update, text)
		})
	}, filters.And(filters.ChatID(cfg.LogChatID), isApkFile())))

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

func isApkFile() filters.FilterOperand {
	return func(df *filters.DataFilter) bool {
		message, ok := df.RawUpdate.(tgTypes.Message)
		if !ok || message.Document == nil {
			return false
		}
		return strings.HasSuffix(strings.ToLower(message.Document.FileName), ".apk") ||
			message.Document.MimeType == "application/vnd.android.package-archive"
	}
}

func isPatchCaption(caption string) bool {
	fields := strings.Fields(caption)
	if len(fields) == 0 {
		return false
	}
	command, _, _ := strings.Cut(strings.ToLower(fields[0]), "@")
	return strings.TrimLeft(command, strings.Join(supportedAliases, "")) == "patch"
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

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/gemini"
	"github.com/TGScheme/TLExtractorBot/internal/github"
	"github.com/TGScheme/TLExtractorBot/internal/services"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph"
	"github.com/robfig/cron/v3"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		gologging.Fatal(err)
	}

	database, err := db.NewDB(cfg)
	if err != nil {
		gologging.Fatal(err)
	}

	settings, err := database.SettingsStore.GetSettings()
	if err != nil {
		gologging.Fatal(err)
	}

	botClient, err := bot.New(cfg)
	if err != nil {
		gologging.Fatal(err)
	}
	defer botClient.Close()
	githubClient, err := github.New(cfg)
	if err != nil {
		gologging.Fatal(err)
	}
	geminiClient, err := gemini.New(cfg, settings.LlmModel)
	if err != nil {
		gologging.Fatal(err)
	}
	telegraphClient, err := telegraph.New(cfg)
	if err != nil {
		gologging.Fatal(err)
	}

	service := services.New(
		cfg, database, botClient, githubClient,
		geminiClient, telegraphClient, scheme.New(database),
	)

	c := cron.New()
	if err = service.Register(c); err != nil {
		gologging.Fatal(err)
	}
	c.Start()
	defer c.Stop()

	services.RegisterCommands(cfg, database, botClient, geminiClient, service)
	botClient.UpdateUptime(true, "")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	botClient.UpdateUptime(false, "shutdown")
}

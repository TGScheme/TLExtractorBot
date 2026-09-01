package services

import (
	"sync/atomic"

	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/gemini"
	"github.com/TGScheme/TLExtractorBot/internal/github"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/bot"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph"
	"github.com/robfig/cron/v3"
)

type Service struct {
	cfg       *config.Config
	db        *db.DB
	bot       *bot.Client
	github    *github.Client
	gemini    *gemini.Client
	telegraph *telegraph.Client
	scheme    *scheme.Client

	building atomic.Bool
	patch    atomic.Bool
}

func New(
	cfg *config.Config,
	database *db.DB,
	botClient *bot.Client,
	githubClient *github.Client,
	geminiClient *gemini.Client,
	telegraphClient *telegraph.Client,
	schemeClient *scheme.Client,
) *Service {
	return &Service{
		cfg:       cfg,
		db:        database,
		bot:       botClient,
		github:    githubClient,
		gemini:    geminiClient,
		telegraph: telegraphClient,
		scheme:    schemeClient,
	}
}

func (s *Service) Register(c *cron.Cron) error {
	if _, err := c.AddFunc("@every 30s", s.pollCoreFork); err != nil {
		return err
	}
	if _, err := c.AddFunc("@every 10s", s.pollSources); err != nil {
		return err
	}
	if _, err := c.AddFunc("0 */6 * * *", s.backupDatabase); err != nil {
		return err
	}
	return nil
}

func (s *Service) RequestPatch() bool {
	if s.building.Load() {
		return false
	}
	s.patch.Store(true)
	return true
}

func (s *Service) IsBuilding() bool { return s.building.Load() }

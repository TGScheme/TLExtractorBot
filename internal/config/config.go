package config

import (
	"fmt"
	"os"
	"strconv"
)

func Load() (*Config, error) {
	cfg := &Config{
		BotToken:       os.Getenv("BOT_TOKEN"),
		APIID:          int(getEnvInt64("API_ID", 0)),
		APIHash:        os.Getenv("API_HASH"),
		ChannelID:      getEnvInt64("CHANNEL_ID", 0),
		LogChatID:      getEnvInt64("LOG_CHAT_ID", 0),
		BackupChatID:   getEnvInt64("BACKUP_CHAT_ID", 0),
		TelegraphToken: os.Getenv("TELEGRAPH_TOKEN"),

		GitHubAppID:          getEnvInt64("GITHUB_APP_ID", 0),
		GitHubInstallationID: getEnvInt64("GITHUB_INSTALLATION_ID", 0),
		GitHubPrivateKeyPath: getEnvString("GITHUB_PRIVATE_KEY_PATH", "/run/secrets/github.pem"),

		GeminiToken: os.Getenv("GEMINI_TOKEN"),

		SchemeRepoOwner: getEnvString("SCHEME_REPO_OWNER", "TGScheme"),
		SchemeRepoName:  getEnvString("SCHEME_REPO_NAME", "Schema"),

		BannerURL: getEnvString("BANNER_URL", "https://telegra.ph/file/5d6793d8428d3ce93fd95.png"),

		BannersRepoOwner: getEnvString("BANNERS_REPO_OWNER", "TGScheme"),
		BannersRepoName:  getEnvString("BANNERS_REPO_NAME", "Banners"),

		WorkDir:    getEnvString("WORK_DIR", "/var/lib/tlextractor"),
		JadxJar:    getEnvString("JADX_JAR", "/opt/jadx/lib/jadx-1.5.0-all.jar"),
		ExtractJar: getEnvString("JADX_EXTRACT_JAR", "/opt/jadx/lib/tlextract.jar"),
		JavaBin:    getEnvString("JAVA_BIN", "java"),

		JadxThreads: int(getEnvInt64("JADX_THREADS", 0)),
		JadxJVMOpts: getEnvString("JADX_JVM_OPTS", "-Xms256M -XX:MaxRAMPercentage=70.0"),

		DBHost:     getEnvString("DB_HOST", "db"),
		DBPort:     int(getEnvInt64("DB_PORT", 5432)),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),

		ValkeyAddr: getEnvString("VALKEY_ADDR", "valkey:6379"),
	}

	for _, required := range []struct {
		name  string
		empty bool
	}{
		{"BOT_TOKEN", cfg.BotToken == ""},
		{"API_ID", cfg.APIID == 0},
		{"API_HASH", cfg.APIHash == ""},
		{"TELEGRAPH_TOKEN", cfg.TelegraphToken == ""},
		{"GEMINI_TOKEN", cfg.GeminiToken == ""},
		{"CHANNEL_ID", cfg.ChannelID == 0},
		{"LOG_CHAT_ID", cfg.LogChatID == 0},
		{"BACKUP_CHAT_ID", cfg.BackupChatID == 0},
		{"GITHUB_APP_ID", cfg.GitHubAppID == 0},
		{"GITHUB_INSTALLATION_ID", cfg.GitHubInstallationID == 0},
		{"DB_USER", cfg.DBUser == ""},
		{"DB_PASSWORD", cfg.DBPassword == ""},
		{"DB_NAME", cfg.DBName == ""},
	} {
		if required.empty {
			return nil, fmt.Errorf("%s environment variable not set", required.name)
		}
	}
	if _, err := os.Stat(cfg.GitHubPrivateKeyPath); err != nil {
		return nil, fmt.Errorf("github private key: %w", err)
	}
	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnvString(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return fallback
}

package config

type Config struct {
	Debug bool

	BotToken       string
	APIID          int
	APIHash        string
	ChannelID      int64
	LogChatID      int64
	BackupChatID   int64
	TelegraphToken string

	GitHubAppID          int64
	GitHubInstallationID int64
	GitHubPrivateKeyPath string

	GeminiToken string

	SchemeRepoOwner string
	SchemeRepoName  string

	BannerURL string

	BannersRepoOwner string
	BannersRepoName  string

	WorkDir    string
	JadxJar    string
	ExtractJar string
	JavaBin    string

	JadxThreads int
	JadxJVMOpts string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	ValkeyAddr string
}

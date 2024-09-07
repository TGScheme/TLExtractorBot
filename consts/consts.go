package consts

import (
	"TLExtractor/utils/package_manager/types"
	"errors"
	"path"
	"regexp"
	"time"
)

// Api Links
const (
	//TDesktopTL   = "https://raw.githubusercontent.com/telegramdesktop/tdesktop/65f7bdb91411e5f03fb97df7b5f1f5b3adb01069/Telegram/SourceFiles/mtproto/scheme/api.tl"
	//AppCenterApi = "https://install.appcenter.ms/api/v0.1/apps/%s/%s/distribution_groups/%s/releases/961"
	TDesktopTL   = "https://raw.githubusercontent.com/telegramdesktop/tdesktop/%s/Telegram/SourceFiles/mtproto/scheme/api.tl"
	AppCenterApi = "https://install.appcenter.ms/api/v0.1/apps/%s/%s/distribution_groups/%s/%s"
	E2ETL        = "https://core.telegram.org/schema/end-to-end-json"
	TelegraphApi = "https://api.telegra.ph"
	TelegraphUrl = "https://telegra.ph"
	GithubURL    = "https://github.com"
)

var (
	TDesktopBranch          = "dev"
	AppCenterAndroidRelease = "latest"
)

// Constants
const (
	Developer          = "drklo-2kb-ghpo"
	AppName            = "telegram-beta-2"
	Distribution       = "all-users-of-telegram-beta-2"
	ServiceDisplayName = "TL Extractor Service"
	ServiceDescription = "Automatically fetches, decompile and commits new Telegram Android TL schema changes."
	ServiceName        = "tl-extractor"
	UpdateMessageRate  = time.Second * 3
	CheckInterval      = time.Second * 1
)

// Github
var (
	SchemeRepoOwner = "TGScheme"
	SchemeRepoName  = "Schema"
)

// Paths
var (
	CredentialsFolder = "credentials.json"
	StorageFolder     = "storage.json"
	PackagesFolder    = "packages"
	TempFolder        = "temp"
	GithubPem         = "github.pem"
	TempBins          = path.Join(TempFolder, "bins")
	TempPackages      = path.Join(TempFolder, "packages")
	TempApk           = path.Join(TempBins, "telegram.apk")
	TempDecompiled    = path.Join(TempFolder, "decompiled")
	TempSources       = path.Join(TempDecompiled, "sources", "org", "telegram", "tgnet")
)

var Requirements = []types.RequireInfo{
	{
		Package: "skylot/jadx",
		File:    "jadx-[0-9.]+\\.zip",
	},
	{
		Package:     "skylot/jadx/jadx-gui",
		File:        "jadx-gui-[0-9.]+-with-jre-win\\.zip",
		OnlyWindows: true,
	},
}

// Regular Expressions
var (
	TLSchemeLineRgx = regexp.MustCompile(`(\S+)#(\w+) *({\S+})? *#* *\[* *([^}=\]]*) *]* = ([^;]+)`)
	OldLayers       = []*regexp.Regexp{
		regexp.MustCompile(`Old[0-9]*$`),
		regexp.MustCompile(`ToBeDeprecated$`),
		regexp.MustCompile(`^\S+[^0-9p][0-9]$`),
		regexp.MustCompile(`^TL\.FileEncryptedLocation$`),
		regexp.MustCompile(`^TL\.DocumentEncrypted$`),
		regexp.MustCompile(`ToDelete$`),
		regexp.MustCompile(`^TL\.MessageEncryptedAction$`),
		regexp.MustCompile(`^TL_message\.Secret$`),
		regexp.MustCompile(`^secret$`),
		regexp.MustCompile(`Layer[0-9]+$`),
	}
	BrokenNames = map[*regexp.Regexp]string{
		regexp.MustCompile(`^((?P<first>is_admin)|is_(?P<second>.*))$`): "$first$second",
		regexp.MustCompile(`^web_`):                                     "",
		regexp.MustCompile(`__b`):                                       "_B",
		regexp.MustCompile(`_item$`):                                    "",
		regexp.MustCompile(`^hash2$`):                                   "hash",
		regexp.MustCompile(`^via_invite$`):                              "via_request",
		regexp.MustCompile(`^_`):                                        "",
		regexp.MustCompile(`^doc$`):                                     "id",
	}
)

var SupportedBotAliases = []string{
	".",
	"/",
	"!",
}

// MtProto Errors
var (
	ConstructorNotFound = errors.New("constructor not found")
	NotTLRPC            = errors.New("not TLRPC")
	OldLayer            = errors.New("old layer")
	UnknownType         = errors.New("unknown type")
	FlagNotFound        = errors.New("flag not found")
)

// Generic Errors
var (
	PackageNotFound = errors.New("package not found")
	JadxNotFound    = errors.New("jadx not found")
	InvalidToken    = errors.New("invalid token")
)

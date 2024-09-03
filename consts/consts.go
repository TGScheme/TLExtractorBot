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
	TDesktopSources  = "https://raw.githubusercontent.com/telegramdesktop/tdesktop/%s/Telegram/SourceFiles"
	TDLibSources     = "https://raw.githubusercontent.com/tdlib/td/master/"
	TDesktopTL       = TDesktopSources + "/mtproto/scheme/api.tl"
	TDLibTL          = TDLibSources + "td/generate/scheme/telegram_api.tl"
	TDAndroidBetaAPI = "https://telegram.org/dl/android/apk-public-beta.json"
	E2ETL            = "https://core.telegram.org/schema/end-to-end-json"
	TelegraphApi     = "https://api.telegra.ph"
	TelegraphUrl     = "https://telegra.ph"
	GithubURL        = "https://github.com"
	MainReleasedTL   = "https://corefork.telegram.org"
)

var TDesktopBranch = "dev"

// Constants
const (
	ServiceDisplayName = "TL Extractor Service"
	ServiceDescription = "Automatically fetches, decompile and commits new Telegram Android TL schema changes."
	ServiceName        = "tl-extractor"
	UpdateMessageRate  = time.Second * 3
	MaxGithubRequests  = 5000 - 100 // 100 is a reserved amount
	NumSources         = 3
	IOSMtProtoPath     = "Payload/Telegram.app/Frameworks/TelegramCoreFramework.framework/TelegramCoreFramework"
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
		Package:     "skylot/jadx",
		File:        "jadx-[0-9.]+\\.zip",
		VersionLock: "v1.5.0",
	},
	{
		Package:     "skylot/jadx/jadx-gui",
		File:        "jadx-gui-[0-9.]+-with-jre-win\\.zip",
		OnlyWindows: true,
		VersionLock: "v1.5.0",
	},
}

// Regular Expressions
var (
	TLSchemeLineRgx     = regexp.MustCompile(`(\S+)#(\w+) *({\S+})? *#* *\[* *([^}=\]]*) *]* = ([^;]+)`)
	TDeskVersionRgx     = regexp.MustCompile(`AppVersion *?= *?([0-9]+);`)
	TDLibVersionRgx     = regexp.MustCompile(`project\(TDLib\s+VERSION\s+([0-9.]+)`)
	TDLibLayerRgx       = regexp.MustCompile(`constexpr int32 MTPROTO_LAYER = ([0-9]+);`)
	TDeskVersionNameRgx = regexp.MustCompile(`AppVersionStr *?= *?"([0-9.]+)";`)
	OldLayers           = []*regexp.Regexp{
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
		regexp.MustCompile(`^TL_messages\.SendEncryptedMultiMedia$`),
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
	BrokenTypes = map[*regexp.Regexp]string{
		regexp.MustCompile(`^InputChatlistDialogFilter$`): "InputChatlist",
		regexp.MustCompile(`PaymentSavedCredentialsCard`): "PaymentSavedCredentials",
	}
	UnusedTypes = []string{
		"ipPortSecret",
		"ipPort",
		"accessPointRule",
		"help.configSimple",
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

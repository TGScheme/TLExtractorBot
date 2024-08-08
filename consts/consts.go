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
	TDesktopTL   = "https://raw.githubusercontent.com/telegramdesktop/tdesktop/dev/Telegram/SourceFiles/mtproto/scheme/api.tl"
	AppCenterApi = "https://install.appcenter.ms/api/v0.1/apps/%s/%s/distribution_groups/%s/releases/latest"
	E2ETL        = "https://core.telegram.org/schema/end-to-end-json"
	TelegraphApi = "https://api.telegra.ph"
	TelegraphUrl = "https://telegra.ph"
	GithubURL    = "https://github.com"
)

// Constants
const (
	Developer         = "drklo-2kb-ghpo"
	AppName           = "telegram-beta-2"
	Distribution      = "all-users-of-telegram-beta-2"
	UpdateMessageRate = time.Second * 3
	CheckInterval     = time.Second * 1
	ServiceName       = "tl-extractor-service"
)

// Github
var (
	SchemeRepoOwner = "TGScheme"
	SchemeRepoName  = "Schema"
)

// Paths
var (
	EnvFolder         = ".env"
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

// Resources
var (
	Templates map[string]string
	Resources map[string][]byte
)

// Regular Expressions
var (
	TLSchemeLineRgx = regexp.MustCompile("(\\S+)#(\\w+) *({\\S+})? *#* *\\[* *([^}=\\]]*) *]* = ([^;]+)")
	OldLayers       = []*regexp.Regexp{
		regexp.MustCompile("Old[0-9]*$"),
		regexp.MustCompile("ToBeDeprecated$"),
		regexp.MustCompile("^\\S+[^0-9p][0-9]$"),
		regexp.MustCompile("^TL\\.FileEncryptedLocation$"),
		regexp.MustCompile("^TL\\.DocumentEncrypted$"),
		regexp.MustCompile("ToDelete$"),
		regexp.MustCompile("^TL\\.MessageEncryptedAction$"),
		regexp.MustCompile("^TL_message\\.Secret$"),
		regexp.MustCompile("^secret$"),
		regexp.MustCompile("Layer[0-9]+$"),
	}
)

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

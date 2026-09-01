package consts

import (
	"path"
	"regexp"
	"time"
)

const (
	TDesktopSources = "https://raw.githubusercontent.com/telegramdesktop/tdesktop/%s/Telegram/SourceFiles"
	TDLibSources    = "https://raw.githubusercontent.com/tdlib/td/master/"
	TDesktopTL      = TDesktopSources + "/mtproto/scheme/api.tl"
	TDLibTL         = TDLibSources + "td/generate/scheme/telegram_api.tl"
	E2ETL           = "https://core.telegram.org/schema/end-to-end-json"
	TelegraphApi    = "https://api.telegra.ph"
	TelegraphUrl    = "https://telegra.ph"
	GithubURL       = "https://github.com"
	MainReleasedTL  = "https://corefork.telegram.org"
)

var TDesktopBranch = "dev"

const (
	UpdateMessageRate     = time.Second * 3
	AndroidBetaChannel    = "TAndroidBeta"
	ChannelPostWindow     = 20
	DownloadThreads       = 8
	MTProtoSessionFile    = "mtproto.session"
	MTProtoReconnectDelay = time.Second * 15
)

var (
	SchemeRepoOwner = "TGScheme"
	SchemeRepoName  = "Schema"
)

var (
	TempFolder      = "temp"
	TgnetPackage    = "org.telegram.tgnet"
	TempBins        = path.Join(TempFolder, "bins")
	TempApk         = path.Join(TempBins, "telegram.apk")
	TempDecompiled  = path.Join(TempFolder, "decompiled")
	TempSourcesRoot = path.Join(TempDecompiled, "sources")
	TempSources     = path.Join(TempSourcesRoot, "org", "telegram", "tgnet")
)

var (
	TLSchemeLineRgx     = regexp.MustCompile(`(\S+)#(\w+) *({\S+})? *#* *\[* *([^}=\]]*) *]* = ([^;]+)`)
	TDeskVersionRgx     = regexp.MustCompile(`AppVersion *?= *?([0-9]+);`)
	TDLibVersionRgx     = regexp.MustCompile(`project\(TDLib\s+VERSION\s+([0-9.]+)`)
	TDLibLayerRgx       = regexp.MustCompile(`constexpr int32 MTPROTO_LAYER = ([0-9]+);`)
	TDeskVersionNameRgx = regexp.MustCompile(`AppVersionStr *?= *?"([0-9.]+)";`)
	DigitVersionRgx     = regexp.MustCompile(`^\S+[^0-9p][0-9]+$`)
	BetaPostVersionRgx  = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)+)\s*\(([0-9]+)\)`)
	OldLayers           = []*regexp.Regexp{
		regexp.MustCompile(`Old[0-9]*$`),
		regexp.MustCompile(`ToBeDeprecated$`),
		regexp.MustCompile(`(?i)ToDelete$`),
		regexp.MustCompile(`Layer[0-9]+$`),
		regexp.MustCompile(`(?i)_legacy$`),
		regexp.MustCompile(`(?i)^TL_contactLink`),
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

const (
	JavaClassQuery = `
		(class_declaration
			(identifier)
			(superclass
				[
					(type_identifier)
					(scoped_type_identifier)
				]
			)
		) @class_declaration
	`
	JavaClassWithNameQuery = `
		(class_declaration
			((identifier) @class_name)
	    	(superclass
				[
					(type_identifier)
					(scoped_type_identifier)
					(generic_type)
				] @class_extends
			)
	    ) @class_declaration
	`
	ExtractJavaVars = `
		(field_declaration
        	([
				(array_type)
                (integral_type)
				(floating_point_type)
				(boolean_type)
				(generic_type)
				(type_identifier)
				(scoped_type_identifier)
			] @var_type)
            (variable_declarator
				((identifier) @var_name)
                ([
					(decimal_integer_literal)
                    (unary_expression)
				] @var_value)?
            )
        )
	`
	ExtractJavaFunctions = `
		(method_declaration
			((identifier) @function_name)
            ((block) @function_body)
		) @function_declaration
	`
)

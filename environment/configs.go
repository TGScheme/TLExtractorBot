package environment

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
	"os"
	"path"
)

var LocalStorage storage
var CredentialsStorage credentials

type storage struct {
	LastVersionCode   uint32                                        `json:"last_version_code"`
	LastTDeskID       int                                           `json:"last_tdesk_id"`
	LastTDLibID       int                                           `json:"last_tdlib_id"`
	ChannelID         int64                                         `json:"channel_id"`
	LogChatID         int64                                         `json:"log_chat_id"`
	LastCoreForkLayer int                                           `json:"last_corefork_layer"`
	MessageId         int64                                         `json:"message_id"`
	StableLayer       *types.TLFullScheme                           `json:"stable_layer,omitempty"`
	PreviewLayer      *types.TLFullScheme                           `json:"preview_layer,omitempty"`
	UpstreamLayer     *types.TLFullScheme                           `json:"upstream_layer,omitempty"`
	PatchedObjects    map[types.PatchOS]map[string]*types.PatchInfo `json:"patched_objects"`
	BannerURL         string                                        `json:"banner_url"`
	RecentLayers      []int                                         `json:"recent_layers"`
	ReleasedLayers    map[int]types.ReleasedLayer                   `json:"released_layers"`
}

func (c storage) Commit() {
	commit(path.Join(EnvFolder, consts.StorageFolder), c)
}

type credentials struct {
	BotToken       string `json:"bot_token"`
	TelegraphToken string `json:"telegraph_token"`
	ApplicationID  int    `json:"application_id"`
	InstallationID int    `json:"installation_id"`
	GeminiToken    string `json:"gemini_token"`
	LLMModel       string `json:"llm_model"`
}

func (c credentials) Commit() {
	commit(path.Join(EnvFolder, consts.CredentialsFolder), c)
}

func commit(path string, data any) {
	marshal, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(path, marshal, os.ModePerm); err != nil {
		panic(err)
	}
}

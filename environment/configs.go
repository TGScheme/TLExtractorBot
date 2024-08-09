package environment

import (
	"TLExtractor/consts"
	"TLExtractor/logging"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
	"os"
	"path"
)

var LocalStorage storage
var CredentialsStorage credentials

type storage struct {
	LastID       uint16              `json:"last_id"`
	ChannelID    int64               `json:"channel_id"`
	MessageId    int64               `json:"message_id"`
	StableLayer  *types.TLFullScheme `json:"stable_layer,omitempty"`
	PreviewLayer *types.TLFullScheme `json:"preview_layer,omitempty"`
	BannerURL    string              `json:"banner_url"`
	RecentLayers []int               `json:"recent_layers"`
	BotName      string              `json:"bot_name"`
}

func (c storage) Commit() {
	commit(path.Join(consts.EnvFolder, consts.StorageFolder), c)
}

type credentials struct {
	BotToken       string `json:"bot_token"`
	TelegraphToken string `json:"telegraph_token"`
	ApplicationID  int64  `json:"application_id"`
	InstallationID int64  `json:"installation_id"`
}

func (c credentials) Commit() {
	commit(path.Join(consts.EnvFolder, consts.CredentialsFolder), c)
}

func commit(path string, data any) {
	marshal, err := json.Marshal(data)
	if err != nil {
		logging.Fatal(err)
	}
	if err = os.WriteFile(path, marshal, os.ModePerm); err != nil {
		logging.Fatal(err)
	}
}

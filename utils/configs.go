package utils

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
	LastID       uint16              `json:"last_id"`
	ChannelID    int64               `json:"channel_id"`
	MessageId    int64               `json:"message_id"`
	StableLayer  *types.TLFullScheme `json:"stable_layer,omitempty"`
	PreviewLayer *types.TLFullScheme `json:"preview_layer,omitempty"`
	BannerURL    string              `json:"banner_url"`
	RecentLayers []int               `json:"recent_layers"`
	ScreenPid    string              `json:"screen_name"`
}

func (c storage) Commit() error {
	return commit(path.Join(consts.EnvFolder, consts.StorageFolder), c)
}

type credentials struct {
	BotToken       string `json:"bot_token"`
	TelegraphToken string `json:"telegraph_token"`
	ApplicationID  int64  `json:"application_id"`
	InstallationID int64  `json:"installation_id"`
}

func (c credentials) Commit() error {
	return commit(path.Join(consts.EnvFolder, consts.CredentialsFolder), c)
}

func commit(path string, data any) error {
	marshal, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, marshal, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func LoadConfigs() error {
	if err := os.MkdirAll(consts.EnvFolder, os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	file, _ := os.ReadFile(path.Join(consts.EnvFolder, consts.StorageFolder))
	_ = json.Unmarshal(file, &LocalStorage)
	file, _ = os.ReadFile(path.Join(consts.EnvFolder, consts.CredentialsFolder))
	_ = json.Unmarshal(file, &CredentialsStorage)
	return nil
}

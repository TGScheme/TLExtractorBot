package utils

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
	"os"
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
}

func (c storage) Commit() error {
	return commit(consts.StorageFolder, c)
}

type credentials struct {
	BotToken       string `json:"bot_token"`
	TelegraphToken string `json:"telegraph_token"`
	ApplicationID  int64  `json:"application_id"`
	InstallationID int64  `json:"installation_id"`
}

func (c credentials) Commit() error {
	return commit(consts.CredentialsFolder, c)
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
	file, _ := os.ReadFile(consts.StorageFolder)
	_ = json.Unmarshal(file, &LocalStorage)
	file, _ = os.ReadFile(consts.CredentialsFolder)
	_ = json.Unmarshal(file, &CredentialsStorage)
	return nil
}

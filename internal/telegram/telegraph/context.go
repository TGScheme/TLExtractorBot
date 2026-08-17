package telegraph

import (
	"encoding/json"
	"fmt"

	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
)

type Client struct {
	accessToken string
	accountInfo types.AccountInfo
}

func New(cfg *config.Config) (*Client, error) {
	res, err := http.ExecuteRequest(
		fmt.Sprintf("%s/getAccountInfo?access_token=%s&fields=[\"short_name\",\"author_name\",\"author_url\"]",
			consts.TelegraphApi, cfg.TelegraphToken),
	)
	if err != nil {
		return nil, err
	}
	var info types.AccountInfo
	if err = json.Unmarshal(res.Body, &info); err != nil {
		return nil, err
	}
	if !info.OK {
		return nil, fmt.Errorf("invalid telegraph token")
	}
	return &Client{accessToken: cfg.TelegraphToken, accountInfo: info}, nil
}

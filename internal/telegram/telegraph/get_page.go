package telegraph

import (
	"encoding/json"
	"fmt"

	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
)

func (ctx *Client) PageTitle(path string) (string, error) {
	res, err := http.ExecuteRequest(
		fmt.Sprintf("%s/getPage/%s?access_token=%s", consts.TelegraphApi, path, ctx.accessToken),
	)
	if err != nil {
		return "", err
	}
	var page struct {
		OK     bool `json:"ok"`
		Result struct {
			Title string `json:"title"`
		} `json:"result"`
	}
	if err = json.Unmarshal(res.Body, &page); err != nil {
		return "", err
	}
	if !page.OK {
		return "", fmt.Errorf("failed to read page %s: %s", path, string(res.Body))
	}
	return page.Result.Title, nil
}

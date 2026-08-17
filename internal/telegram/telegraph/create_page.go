package telegraph

import (
	"encoding/json"
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"

	"github.com/Laky-64/http"
)

func (ctx *Client) CreatePage(title string, html string) (string, error) {
	dom, err := parseHtml(html)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(
		types.CreatePageRequest{
			AuthorName:  ctx.accountInfo.Result.AuthorName,
			AuthorURL:   ctx.accountInfo.Result.AuthorURL,
			AccessToken: ctx.accessToken,
			Title:       title,
			Content:     dom,
		},
	)
	if err != nil {
		return "", err
	}
	res, err := http.ExecuteRequest(
		fmt.Sprintf("%s/createPage", consts.TelegraphApi),
		http.Method("POST"),
		http.Headers(map[string]string{"Content-Type": "application/json"}),
		http.Body(body),
	)
	if err != nil {
		return "", err
	}
	var createRes types.CreatePageResult
	err = json.Unmarshal(res.Body, &createRes)
	if err != nil {
		return "", err
	}
	if !createRes.OK {
		return "", fmt.Errorf("failed to create page: %s", string(res.Body))
	}
	return createRes.Result.URL, nil
}

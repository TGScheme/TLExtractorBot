package telegraph

import (
	"encoding/json"
	"fmt"

	"github.com/Laky-64/http"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
)

func (ctx *Client) savePage(endpoint, path, title, html string) (types.PageInfo, error) {
	dom, err := parseHtml(html)
	if err != nil {
		return types.PageInfo{}, err
	}
	body, err := json.Marshal(
		types.CreatePageRequest{
			AuthorName:  ctx.accountInfo.Result.AuthorName,
			AuthorURL:   ctx.accountInfo.Result.AuthorURL,
			AccessToken: ctx.accessToken,
			Path:        path,
			Title:       title,
			Content:     dom,
		},
	)
	if err != nil {
		return types.PageInfo{}, err
	}
	res, err := http.ExecuteRequest(
		endpoint,
		http.Method("POST"),
		http.Headers(map[string]string{"Content-Type": "application/json"}),
		http.Body(body),
	)
	if err != nil {
		return types.PageInfo{}, err
	}
	var pageRes types.CreatePageResult
	err = json.Unmarshal(res.Body, &pageRes)
	if err != nil {
		return types.PageInfo{}, err
	}
	if !pageRes.OK {
		return types.PageInfo{}, fmt.Errorf("failed to save page: %s", string(res.Body))
	}
	return types.PageInfo{Path: pageRes.Result.Path, URL: pageRes.Result.URL}, nil
}

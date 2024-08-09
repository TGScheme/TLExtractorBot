package telegraph

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/http"
	"TLExtractor/telegram/telegraph/types"
	"encoding/json"
	"fmt"
)

func (ctx *context) CreatePage(title string, html string) (string, error) {
	dom, err := parseHtml(html)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(
		types.CreatePageRequest{
			AuthorName:  ctx.accountInfo.Result.AuthorName,
			AuthorURL:   ctx.accountInfo.Result.AuthorURL,
			AccessToken: environment.CredentialsStorage.TelegraphToken,
			Title:       title,
			Content:     dom,
		},
	)
	if err != nil {
		return "", err
	}
	res := http.ExecuteRequest(
		fmt.Sprintf("%s/createPage", consts.TelegraphApi),
		http.Method("POST"),
		http.Headers(map[string]string{"Content-Type": "application/json"}),
		http.Body(body),
	)
	if res.Error != nil {
		return "", res.Error
	}
	var createRes types.CreatePageResult
	err = json.Unmarshal(res.Read(), &createRes)
	if err != nil {
		return "", err
	}
	if !createRes.OK {
		return "", fmt.Errorf("failed to create page: %s", string(res.Read()))
	}
	return createRes.Result.URL, nil
}

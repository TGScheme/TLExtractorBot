package telegraph

import (
	"fmt"

	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
)

func (ctx *Client) CreatePage(title string, html string) (types.PageInfo, error) {
	return ctx.savePage(fmt.Sprintf("%s/createPage", consts.TelegraphApi), "", title, html)
}

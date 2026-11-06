package telegraph

import (
	"fmt"

	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
)

func (ctx *Client) EditPage(path, title, html string) (types.PageInfo, error) {
	return ctx.savePage(fmt.Sprintf("%s/editPage", consts.TelegraphApi), path, title, html)
}

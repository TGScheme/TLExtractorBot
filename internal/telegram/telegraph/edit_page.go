package telegraph

import (
	"fmt"

	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/telegraph/types"
)

func (ctx *Client) EditPage(path string, title string, html string) (types.PageInfo, error) {
	return ctx.savePage(fmt.Sprintf("%s/editPage/%s", consts.TelegraphApi, path), path, title, html)
}

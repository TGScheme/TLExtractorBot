package scheme

import (
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"

	"github.com/Laky-64/http"
)

func GetScheme(branch string) (*types.TLRemoteScheme, error) {
	res, err := http.ExecuteRequest(fmt.Sprintf(consts.TDesktopTL, branch))
	if err != nil {
		return nil, err
	}
	return ParseTLScheme(res.String())
}

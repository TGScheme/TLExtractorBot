package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"fmt"
	"github.com/Laky-64/http"
)

func GetScheme() (*types.TLRemoteScheme, error) {
	res, err := http.ExecuteRequest(fmt.Sprintf(consts.TDesktopTL, consts.TDesktopBranch))
	if err != nil {
		return nil, err
	}
	return ParseTLScheme(res.String())
}

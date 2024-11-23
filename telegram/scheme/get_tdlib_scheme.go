package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"github.com/Laky-64/http"
	"strconv"
)

func GetTDLibScheme() (*types.TLRemoteScheme, error) {
	res, err := http.ExecuteRequest(consts.TDLibTL)
	if err != nil {
		return nil, err
	}
	rawScheme, err := ParseTLScheme(res.String())
	if err != nil {
		return nil, err
	}
	res, err = http.ExecuteRequest(consts.TDLibSources + "td/telegram/Version.h")
	if err != nil {
		return nil, err
	}
	layerVersion, err := strconv.Atoi(consts.TDLibLayerRgx.FindAllStringSubmatch(res.String(), -1)[0][1])
	if err != nil {
		return nil, err
	}
	rawScheme.Layer = layerVersion
	return rawScheme, nil
}

package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
	"github.com/Laky-64/http"
)

func getE2EScheme() (*types.TLScheme, error) {
	request, err := http.ExecuteRequest(consts.E2ETL)
	if err != nil {
		return nil, err
	}
	var generatedScheme types.TLScheme
	err = json.Unmarshal(request.Body, &generatedScheme)
	if err != nil {
		return nil, err
	}
	return &generatedScheme, nil
}

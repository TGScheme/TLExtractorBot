package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/http"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
)

func getE2EScheme() (*types.TLScheme, error) {
	request := http.ExecuteRequest(consts.E2ETL)
	if request.Error != nil {
		return nil, request.Error
	}
	var generatedScheme types.TLScheme
	err := json.Unmarshal(request.Read(), &generatedScheme)
	if err != nil {
		return nil, err
	}
	return &generatedScheme, nil
}

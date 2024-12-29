package scheme

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme/types"
	"encoding/json"
	"github.com/Laky-64/http"
)

func GetE2EScheme() (*types.TLRemoteScheme, error) {
	request, err := http.ExecuteRequest(consts.E2ETL)
	if err != nil {
		return nil, err
	}
	var generatedScheme types.TLRemoteScheme
	err = json.Unmarshal(request.Body, &generatedScheme)
	if err != nil {
		return nil, err
	}
	for i := range generatedScheme.Constructors {
		generatedScheme.Constructors[i].ForceSecret = true
	}
	for i := range generatedScheme.Methods {
		generatedScheme.Methods[i].ForceSecret = true
	}
	return &generatedScheme, nil
}

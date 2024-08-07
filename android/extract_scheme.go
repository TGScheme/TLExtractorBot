package android

import (
	"TLExtractor/telegram/scheme"
	schemeTypes "TLExtractor/telegram/scheme/types"
)

func ExtractScheme() (*schemeTypes.TLFullScheme, error) {
	rawScheme, err := extractRawScheme()
	if err != nil {
		return nil, err
	}
	mainScheme, err := scheme.MergeOfficial(rawScheme, false)
	if err != nil {
		return nil, err
	}
	e2eScheme, err := scheme.MergeOfficial(rawScheme, true)
	if err != nil {
		return nil, err
	}
	return &schemeTypes.TLFullScheme{
		MainApi: mainScheme.TLScheme,
		E2EApi:  e2eScheme.TLScheme,
		Layer:   mainScheme.Layer,
	}, nil
}

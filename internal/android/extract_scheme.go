package android

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func ExtractScheme(workDir string, client *scheme.Client, branch string) (*schemeTypes.TLFullScheme, []string, error) {
	rawScheme, err := extractRawScheme(workDir)
	if err != nil {
		return nil, nil, err
	}
	extracted := make([]string, 0, len(rawScheme.Constructors)+len(rawScheme.Methods))
	for _, constructor := range rawScheme.Constructors {
		extracted = append(extracted, scheme.ParseConstructor(constructor.Constructor()))
	}
	for _, method := range rawScheme.Methods {
		extracted = append(extracted, scheme.ParseConstructor(method.Constructor()))
	}
	fullScheme, err := client.MergeSmartUpstream(rawScheme, schemeTypes.AndroidPatch, branch)
	if err != nil {
		return nil, nil, err
	}
	return fullScheme, extracted, nil
}

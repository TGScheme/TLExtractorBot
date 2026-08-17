package android

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func ExtractScheme(workDir string, client *scheme.Client, branch string) (*schemeTypes.TLFullScheme, error) {
	rawScheme, err := extractRawScheme(workDir)
	if err != nil {
		return nil, err
	}
	return client.MergeSmartUpstream(rawScheme, schemeTypes.AndroidPatch, branch)
}

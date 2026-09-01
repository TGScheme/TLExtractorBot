package android

import (
	"fmt"
	"time"

	"github.com/Laky-64/gologging"

	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func ExtractScheme(workDir string, client *scheme.Client, branch string) (*schemeTypes.TLFullScheme, []string, error) {
	start := time.Now()
	rawScheme, err := extractRawScheme(workDir)
	if err != nil {
		return nil, nil, err
	}
	gologging.Info(fmt.Sprintf("android: parsed the sources in %s", time.Since(start).Round(time.Millisecond)))
	extracted := make([]string, 0, len(rawScheme.Constructors)+len(rawScheme.Methods))
	for _, constructor := range rawScheme.Constructors {
		extracted = append(extracted, scheme.ParseConstructor(constructor.Constructor()))
	}
	for _, method := range rawScheme.Methods {
		extracted = append(extracted, scheme.ParseConstructor(method.Constructor()))
	}
	start = time.Now()
	fullScheme, err := client.MergeSmartUpstream(rawScheme, schemeTypes.AndroidPatch, branch)
	if err != nil {
		return nil, nil, err
	}
	gologging.Info(fmt.Sprintf("android: merged the upstream in %s", time.Since(start).Round(time.Millisecond)))
	return fullScheme, extracted, nil
}

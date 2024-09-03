package ios

import (
	"TLExtractor/telegram/scheme/types"
	"fmt"
	"github.com/Laky-64/goswift/proxy"
)

func extractRawScheme(file *proxy.Context) (*types.RawTLScheme, error) {
	var scheme types.RawTLScheme
	rawObjects, err := getRawObjects(file)
	if err != nil {
		return nil, err
	}
	fmt.Println(len(rawObjects))
	for _, rawObject := range rawObjects {
		extractObject(file, rawObject)
	}
	_ = rawObjects
	scheme.Layer, err = extractLayer(file)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("not implemented")
}

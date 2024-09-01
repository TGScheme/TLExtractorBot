package scheme

import (
	"TLExtractor/telegram/scheme/types"
	"strings"
)

func filterObjects[T types.TLInterface](objects []T, isE2E bool) []T {
	var filteredObjects []T
	for _, object := range objects {
		isSecretObject := strings.HasPrefix(object.Package(), "decryptedMessage") || object.IsSecret()
		if isSecretObject != isE2E {
			continue
		}
		filteredObjects = append(filteredObjects, object)
	}
	return filteredObjects
}

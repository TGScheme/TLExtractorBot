package types

import (
	"errors"
	"regexp"
)

type ReleaseInfo struct {
	Version string      `json:"name"`
	Assets  []AssetInfo `json:"assets"`
}

func (r ReleaseInfo) GetCompatibleAsset(variant string) (*AssetInfo, error) {
	for _, asset := range r.Assets {
		if regexp.MustCompile(variant).MatchString(asset.Name) {
			return &asset, nil
		}
	}
	return nil, errors.New("no compatible asset found")
}

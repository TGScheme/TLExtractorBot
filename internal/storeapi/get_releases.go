package storeapi

import (
	"github.com/TGScheme/TLExtractorBot/internal/storeapi/types"
)

func GetReleases() ([]types.Release, error) {
	var versions []types.Release
	info, err := GetAppInfo()
	if err != nil {
		return nil, err
	}
	versions = append(versions, types.Release{
		Version:     info.Version,
		VersionCode: info.VersionCode,
	})
	return versions, nil
}

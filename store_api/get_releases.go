package store_api

import (
	"TLExtractor/store_api/types"
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

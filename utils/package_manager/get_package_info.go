package package_manager

import (
	"TLExtractor/github"
	types2 "TLExtractor/utils/package_manager/types"
	"errors"
	"regexp"
)

func getPackageInfo(info types2.RequireInfo) (*types2.PackageInfo, error) {
	release, err := github.Client.GetLastRelease(info.RepoOwner(), info.RepoName(), info.VersionLock)
	if err != nil {
		return nil, err
	}
	var compatibleAsset *types2.AssetInfo
	for _, asset := range release.Assets {
		if regexp.MustCompile(info.File).MatchString(*asset.Name) {
			compatibleAsset = &types2.AssetInfo{
				Name: *asset.Name,
				URL:  *asset.BrowserDownloadURL,
				Size: *asset.Size,
			}
			break
		}
	}
	if compatibleAsset == nil {
		return nil, errors.New("no compatible asset found")
	}
	return &types2.PackageInfo{
		Name:        info.PackageName(),
		Owner:       info.RepoOwner(),
		Version:     *release.Name,
		DownloadURL: compatibleAsset.URL,
		Size:        compatibleAsset.Size,
		FileName:    compatibleAsset.Name,
	}, nil
}

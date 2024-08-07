package package_manager

import (
	"TLExtractor/http"
	types2 "TLExtractor/utils/package_manager/types"
	"encoding/json"
	"fmt"
)

func getPackageInfo(info types2.RequireInfo) (*types2.PackageInfo, error) {
	var releaseInfo types2.ReleaseInfo
	err := json.Unmarshal(
		http.ExecuteRequest(
			fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", info.RepoOwner(), info.RepoName()),
		).Read(),
		&releaseInfo,
	)
	if err != nil {
		return nil, err
	}
	asset, err := releaseInfo.GetCompatibleAsset(info.File)
	if err != nil {
		return nil, err
	}
	return &types2.PackageInfo{
		Name:        info.PackageName(),
		Owner:       info.RepoOwner(),
		Version:     releaseInfo.Version,
		DownloadURL: asset.URL,
		Size:        asset.Size,
		FileName:    asset.Name,
	}, nil
}

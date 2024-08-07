package package_manager

import (
	"TLExtractor/utils/package_manager/types"
)

func comparePackages(local, remote []types.PackageInfo) []types.PackageInfo {
	var toUpdateOrInstall []types.PackageInfo
	for _, r := range remote {
		found := false
		for _, l := range local {
			if l.Name == r.Name {
				if l.Version != r.Version {
					toUpdateOrInstall = append(toUpdateOrInstall, r)
				}
				found = true
				break
			}
		}
		if !found {
			toUpdateOrInstall = append(toUpdateOrInstall, r)
		}
	}
	return toUpdateOrInstall
}

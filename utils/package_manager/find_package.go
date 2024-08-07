package package_manager

import (
	"TLExtractor/consts"
	"TLExtractor/utils/package_manager/types"
)

func FindPackage(name string) (*types.PackageInfo, error) {
	localPackages, err := installedPackages()
	if err != nil {
		return nil, err
	}
	for _, p := range localPackages {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, consts.PackageNotFound
}

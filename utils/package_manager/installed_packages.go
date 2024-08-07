package package_manager

import (
	"TLExtractor/consts"
	"TLExtractor/utils/package_manager/types"
	"os"
	"path"
	"regexp"
)

func installedPackages() ([]types.PackageInfo, error) {
	dir, err := os.ReadDir(path.Join(consts.BasePath, consts.PackagesFolder))
	if err != nil {
		return nil, err
	}
	var packages []types.PackageInfo
	compile := regexp.MustCompile(`(.*?)-([0-9.]+)`)
	for _, d := range dir {
		if d.IsDir() {
			if compile.MatchString(d.Name()) {
				dataInfo := compile.FindAllStringSubmatch(d.Name(), -1)[0][1:]
				packages = append(packages, types.PackageInfo{
					Name:    dataInfo[0],
					Version: dataInfo[1],
					Path:    path.Join(consts.BasePath, consts.PackagesFolder, d.Name()),
				})
			}
		}
	}
	return packages, nil
}

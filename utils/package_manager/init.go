package package_manager

import (
	"TLExtractor/consts"
	"TLExtractor/logging"
	"TLExtractor/utils/package_manager/types"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
)

func init() {
	if err := os.MkdirAll(path.Join(consts.EnvFolder, consts.PackagesFolder), os.ModePerm); err != nil && !os.IsExist(err) {
		logging.Fatal(err)
	}
	var requirements []types.PackageInfo
	for _, p := range consts.Requirements {
		if p.OnlyWindows && runtime.GOOS != "windows" {
			continue
		}
		packageInfo, err := getPackageInfo(p)
		if err != nil {
			logging.Fatal(err)
		}
		requirements = append(requirements, *packageInfo)
	}
	localPackages, err := installedPackages()
	if err != nil {
		logging.Fatal(err)
	}
	downloadPackages := comparePackages(localPackages, requirements)
	if len(downloadPackages) > 0 {
		var missingPackages []string
		for _, p := range downloadPackages {
			missingPackages = append(missingPackages, p.GetFullName())
		}
		for _, p := range downloadPackages {
			fmt.Println(fmt.Sprintf("Collecting %s==%s", p.Name, p.Version))
			if err = download(p); err != nil {
				logging.Fatal(err)
			}
		}
		fmt.Println("Installing collected packages: " + strings.Join(missingPackages, " "))
		for _, p := range downloadPackages {
			if err = install(p); err != nil {
				logging.Fatal(err)
			}
		}
	}
}

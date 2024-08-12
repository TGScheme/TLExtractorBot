package package_manager

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/utils/package_manager/types"
	"fmt"
	"github.com/Laky-64/gologging"
	"os"
	"path"
	"runtime"
	"strings"
)

func CheckPackages() {
	if err := os.MkdirAll(path.Join(environment.EnvFolder, consts.PackagesFolder), os.ModePerm); err != nil && !os.IsExist(err) {
		gologging.Fatal(err)
	}
	var requirements []types.PackageInfo
	for _, p := range consts.Requirements {
		if p.OnlyWindows && runtime.GOOS != "windows" {
			continue
		}
		packageInfo, err := getPackageInfo(p)
		if err != nil {
			gologging.Fatal(err)
		}
		requirements = append(requirements, *packageInfo)
	}
	localPackages, err := installedPackages()
	if err != nil {
		gologging.Fatal(err)
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
				gologging.Fatal(err)
			}
		}
		fmt.Println("Installing collected packages: " + strings.Join(missingPackages, " "))
		for _, p := range downloadPackages {
			if err = install(p); err != nil {
				gologging.Fatal(err)
			}
		}
	}
}

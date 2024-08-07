package package_manager

import (
	"TLExtractor/consts"
	"TLExtractor/utils"
	"TLExtractor/utils/package_manager/types"
	"fmt"
	"os"
	"strings"
)

func CheckPackages(githubClient *github.Context) error {
	if err := os.MkdirAll(consts.PackagesFolder, os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}
	var requirements []types.PackageInfo
	for _, p := range consts.Requirements {
		if p.OnlyWindows && !utils.IsWindows() {
			continue
		}
		packageInfo, err := getPackageInfo(githubClient, p)
		if err != nil {
			return err
		}
		requirements = append(requirements, *packageInfo)
	}
	localPackages, err := installedPackages()
	if err != nil {
		return err
	}
	downloadPackages := comparePackages(localPackages, requirements)
	if len(downloadPackages) > 0 {
		var missingPackages []string
		for _, p := range downloadPackages {
			missingPackages = append(missingPackages, p.GetFullName())
		}
		for _, p := range downloadPackages {
			fmt.Println(fmt.Sprintf("Collecting %s==%s", p.Name, p.Version))
			if err := download(p); err != nil {
				return err
			}
		}
		fmt.Println("Installing collected packages: " + strings.Join(missingPackages, " "))
		for _, p := range downloadPackages {
			if err := install(p); err != nil {
				return err
			}
		}
	}
	return nil
}

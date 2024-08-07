package jadx

import (
	"TLExtractor/utils"
	"TLExtractor/utils/package_manager"
	"path"
)

func javaExec() (string, error) {
	if utils.IsWindows() {
		pkgInfo, err := package_manager.FindPackage("jadx-gui")
		if err != nil {
			return "", err
		}
		return path.Join(pkgInfo.Path, "jre", "bin", "java.exe"), nil
	}
	return "java", nil
}

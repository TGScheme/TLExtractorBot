package jadx

import (
	"TLExtractor/utils/package_manager"
	"path"
	"runtime"
)

func javaExec() (string, error) {
	if runtime.GOOS == "windows" {
		pkgInfo, err := package_manager.FindPackage("jadx-gui")
		if err != nil {
			return "", err
		}
		return path.Join(pkgInfo.Path, "jre", "bin", "java.exe"), nil
	}
	return "java", nil
}

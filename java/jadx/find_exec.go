package jadx

import (
	"TLExtractor/consts"
	"TLExtractor/utils/package_manager"
	"os"
	"path"
	"strings"
)

func findExec() (string, error) {
	pkgInfo, err := package_manager.FindPackage("jadx")
	if err != nil {
		return "", err
	}
	basePath := path.Join(pkgInfo.Path, "lib")
	dir, err := os.ReadDir(basePath)
	if err != nil {
		return "", err
	}
	for _, d := range dir {
		if !d.IsDir() && strings.HasPrefix(d.Name(), "jadx") {
			return path.Join(basePath, d.Name()), nil
		}
	}
	return "", consts.JadxNotFound
}

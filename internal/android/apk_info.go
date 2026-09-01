package android

import (
	"fmt"

	"github.com/shogo82148/androidbinary/apk"
)

type APKInfo struct {
	VersionName string
	VersionCode uint32
}

func ReadAPKInfo(apkPath string) (*APKInfo, error) {
	pkg, err := apk.OpenFile(apkPath)
	if err != nil {
		return nil, fmt.Errorf("open apk: %w", err)
	}
	defer func() {
		_ = pkg.Close()
	}()
	manifest := pkg.Manifest()
	versionName, err := manifest.VersionName.String()
	if err != nil {
		return nil, fmt.Errorf("apk version name: %w", err)
	}
	versionCode, err := manifest.VersionCode.Int32()
	if err != nil {
		return nil, fmt.Errorf("apk version code: %w", err)
	}
	if versionCode <= 0 {
		return nil, fmt.Errorf("apk reports version code %d", versionCode)
	}
	return &APKInfo{VersionName: versionName, VersionCode: uint32(versionCode)}, nil
}

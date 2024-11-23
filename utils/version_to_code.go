package utils

import (
	"math"
	"strconv"
	"strings"
)

func VersionToCode(version string) int {
	rawVersionCode := strings.Split(version, ".")
	versionCode := 0
	for i, v := range rawVersionCode {
		subVersionCode, _ := strconv.Atoi(v)
		versionCode += int(float64(subVersionCode) * math.Pow(100, float64(len(rawVersionCode)-i-1)))
	}
	return versionCode
}

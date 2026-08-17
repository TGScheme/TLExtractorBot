package utils

import "fmt"

const (
	_kB = 1e3
	_MB = 1e6
	_GB = 1e9
	_TB = 1e12
)

func HumanReadableBytes(i int64) (result string) {
	switch {
	case i >= _TB:
		result = fmt.Sprintf("%.02f TB", float64(i)/_TB)
	case i >= _GB:
		result = fmt.Sprintf("%.02f GB", float64(i)/_GB)
	case i >= _MB:
		result = fmt.Sprintf("%.02f MB", float64(i)/_MB)
	case i >= _kB:
		result = fmt.Sprintf("%.02f kB", float64(i)/_kB)
	default:
		result = fmt.Sprintf("%d B", i)
	}
	return
}

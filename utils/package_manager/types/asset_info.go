package types

import "strings"

type AssetInfo struct {
	Name string
	URL  string
	Size int
}

func (a AssetInfo) IsWindows() bool {
	return strings.Contains(a.Name, "-win")
}

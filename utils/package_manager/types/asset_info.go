package types

import "strings"

type AssetInfo struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int    `json:"size"`
}

func (a AssetInfo) IsWindows() bool {
	return strings.Contains(a.Name, "-win")
}

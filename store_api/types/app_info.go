package types

type AppInfo struct {
	Version     string `json:"version"`
	VersionCode uint32 `json:"version_code"`
	FileURL     string `json:"file_url"`
}

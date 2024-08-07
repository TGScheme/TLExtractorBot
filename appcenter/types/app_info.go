package types

type AppInfo struct {
	ID          uint16 `json:"id"`
	VersionName string `json:"short_version"`
	BuildNumber string `json:"version"`
	DownloadURL string `json:"download_url"`
	Size        uint64 `json:"size"`
}

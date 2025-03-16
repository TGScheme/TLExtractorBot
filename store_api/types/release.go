package types

type Release struct {
	Version     string `json:"short_version"`
	VersionCode uint32 `json:"version"`
}

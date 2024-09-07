package types

type Release struct {
	Id           int    `json:"id"`
	ShortVersion string `json:"short_version"`
	Version      string `json:"version"`
}

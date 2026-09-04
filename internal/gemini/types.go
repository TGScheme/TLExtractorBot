package gemini

import (
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

type ChangelogRequest struct {
	Layer       int
	Source      string
	VersionName string
	BuildNumber uint32
	IsPatch     bool
	Scheme      *schemeTypes.TLFullScheme
	Differences *schemeTypes.TLFullDifferences
}

type Changelog struct {
	Title    string             `json:"title"`
	Lead     string             `json:"lead"`
	Sections []ChangelogSection `json:"sections"`
	Items    []ChangelogItem    `json:"items"`

	Descriptions map[string]string `json:"-"`
}

type ChangelogSection struct {
	Title      string   `json:"title"`
	Paragraphs []string `json:"paragraphs"`
	Highlights []string `json:"highlights"`
}

type ChangelogItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

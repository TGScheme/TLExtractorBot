package services

import (
	"encoding/json"
	"maps"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/llm"
)

type layerStory struct {
	Title        string                 `json:"title"`
	Lead         string                 `json:"lead"`
	Sections     []llm.ChangelogSection `json:"sections"`
	Descriptions map[string]string      `json:"descriptions"`
	Model        string                 `json:"model"`
}

func (s *Service) loadStory(layer int) *layerStory {
	raw, err := s.db.PagesStore.GetLayerStory(int64(layer))
	if err != nil || raw == "" {
		return nil
	}
	var story layerStory
	if err = json.Unmarshal([]byte(raw), &story); err != nil {
		gologging.Error("story: unable to read the saved story:", err)
		return nil
	}
	return &story
}

func (s *Service) saveStory(layer int, story *layerStory) {
	raw, err := json.Marshal(story)
	if err != nil {
		gologging.Error("story: unable to encode the story:", err)
		return
	}
	if err = s.db.PagesStore.SetLayerStory(string(raw), int64(layer)); err != nil {
		gologging.Error("story: unable to save the story:", err)
	}
}

func newStory(changelog *llm.Changelog) *layerStory {
	return &layerStory{
		Title:        changelog.Title,
		Lead:         changelog.Lead,
		Sections:     changelog.Sections,
		Descriptions: maps.Clone(changelog.Descriptions),
		Model:        changelog.Model,
	}
}

func (story *layerStory) addDescriptions(changelog *llm.Changelog) {
	if changelog == nil {
		return
	}
	if story.Descriptions == nil {
		story.Descriptions = make(map[string]string)
	}
	for name, description := range changelog.Descriptions {
		if _, ok := story.Descriptions[name]; !ok {
			story.Descriptions[name] = description
		}
	}
}

func (story *layerStory) changelog() *llm.Changelog {
	return &llm.Changelog{
		Title:        story.Title,
		Lead:         story.Lead,
		Sections:     story.Sections,
		Descriptions: story.Descriptions,
		Model:        story.Model,
	}
}

package services

import (
	"fmt"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

const maxRemovalRatio = 0.10

func disappearedObjects(previous, current []string) map[string]bool {
	present := make(map[string]bool, len(current))
	for _, id := range current {
		present[id] = true
	}
	disappeared := make(map[string]bool)
	for _, id := range previous {
		if !present[id] {
			disappeared[id] = true
		}
	}
	return disappeared
}

func (s *Service) applyDisappeared(source string, fullScheme *schemeTypes.TLFullScheme, extracted []string) error {
	previous, err := s.db.LoadExtractedObjects(source)
	if err != nil {
		return err
	}
	disappeared := disappearedObjects(previous, extracted)
	if len(previous) > 0 && float64(len(disappeared)) > float64(len(previous))*maxRemovalRatio {
		return fmt.Errorf(
			"the app dropped %d of %d objects, refusing to publish a likely broken extraction",
			len(disappeared), len(previous),
		)
	}
	if removed := scheme.RemoveObjects(fullScheme, disappeared); removed > 0 {
		gologging.Info(fmt.Sprintf("scheme: removed %d objects dropped by the app", removed))
	}
	return db.IgnoreNoRows(s.db.SaveExtractedObjects(source, extracted))
}

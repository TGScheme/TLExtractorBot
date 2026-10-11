package services

import (
	"fmt"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/db"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

const maxRemovalRatio = 0.10

func resolveDropped(previous, extracted []string, dropped map[string]int64, build int64) (missing, restored []string, excluded map[string]bool) {
	present := make(map[string]bool, len(extracted))
	for _, id := range extracted {
		present[id] = true
	}
	for _, id := range previous {
		if !present[id] {
			missing = append(missing, id)
		}
	}
	excluded = make(map[string]bool, len(dropped)+len(missing))
	for _, id := range missing {
		excluded[id] = true
	}
	for id, droppedAt := range dropped {
		if present[id] && droppedAt < build {
			restored = append(restored, id)
			continue
		}
		excluded[id] = true
	}
	return missing, restored, excluded
}

func (s *Service) applyDropped(update UpdateInfo, fullScheme *schemeTypes.TLFullScheme, extracted []string) error {
	previous, err := s.db.LoadExtractedObjects(update.Package)
	if err != nil {
		return err
	}
	dropped, err := s.db.LoadDroppedObjects()
	if err != nil {
		return err
	}
	build := int64(update.BuildNumber)
	missing, restored, excluded := resolveDropped(previous, extracted, dropped, build)
	if len(previous) > 0 && float64(len(missing)) > float64(len(previous))*maxRemovalRatio {
		return fmt.Errorf(
			"the app dropped %d of %d objects, refusing to publish a likely broken extraction",
			len(missing), len(previous),
		)
	}
	for _, id := range missing {
		if err = s.db.SetDroppedObject(id, build); err != nil {
			return err
		}
	}
	for _, id := range restored {
		if err = s.db.RestoreDroppedObject(id); err != nil {
			return err
		}
	}
	if removed := scheme.RemoveObjects(fullScheme, excluded); removed > 0 {
		gologging.Info(fmt.Sprintf(
			"scheme: kept %d objects dropped by the app out of the scheme (%d newly missing from %s)",
			removed, len(missing), update.Package,
		))
	}
	return db.IgnoreNoRows(s.db.SaveExtractedObjects(update.Package, extracted))
}

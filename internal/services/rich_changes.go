package services

import (
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

type richChange struct {
	Object schemeTypes.TLObjDifferences
	IsE2E  bool
}

func appendRichChanges(changes []richChange, diffs *schemeTypes.TLSchemeDifferences, isE2E bool) []richChange {
	if diffs == nil {
		return changes
	}
	for _, group := range [][]schemeTypes.TLObjDifferences{diffs.MethodsDifference, diffs.ConstructorsDifference} {
		for _, object := range group {
			if !object.IsNew && !object.IsDeleted &&
				len(object.NewFields) == 0 && len(object.RemovedFields) == 0 &&
				len(object.ChangedFields) == 0 && object.ChangedResult == nil {
				continue
			}
			changes = append(changes, richChange{Object: object, IsE2E: isE2E})
		}
	}
	return changes
}

func richChanges(diffs *schemeTypes.TLFullDifferences) []richChange {
	if diffs == nil {
		return nil
	}
	var changes []richChange
	changes = appendRichChanges(changes, diffs.MainApi, false)
	return appendRichChanges(changes, diffs.E2EApi, true)
}

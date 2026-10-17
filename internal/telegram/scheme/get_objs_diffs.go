package scheme

import (
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"slices"
)

func getObjsDiffs[T types.TLInterface](old, new []T, oldLayer, newLayer int) []types.TLObjDifferences {
	var diff []types.TLObjDifferences
	mappedOldObjects := make(map[string]types.TLInterface)
	reverseNames := make(map[string]string)
	for _, oldInterface := range old {
		packageName := diffKey(oldInterface, oldLayer)
		mappedOldObjects[packageName] = oldInterface
		reverseNames[oldInterface.Constructor()] = packageName
	}
	var foundPackages []string
	for _, newInterface := range new {
		packageName := diffKey(newInterface, newLayer)
		if reversedPackageName, ok := reverseNames[newInterface.Constructor()]; ok {
			packageName = reversedPackageName
		}
		foundPackages = append(foundPackages, packageName)
		if _, found := mappedOldObjects[packageName]; !found {
			diff = append(diff, types.TLObjDifferences{
				Object: newInterface,
				IsNew:  true,
			})
		} else {
			var mappedOldFields = make(map[string]string)
			var mappedNewFields = make(map[string]string)
			for _, field := range mappedOldObjects[packageName].Parameters() {
				mappedOldFields[field.Name] = field.Type
			}
			for _, field := range newInterface.Parameters() {
				mappedNewFields[field.Name] = field.Type
			}
			var newFields, removedFields []string
			var changedFields []types.TlDifferentField
			for field, fieldType := range mappedNewFields {
				if _, found = mappedOldFields[field]; !found {
					newFields = append(newFields, field)
				} else if mappedOldFields[field] != fieldType {
					changedFields = append(changedFields, types.TlDifferentField{
						Name:    field,
						OldType: mappedOldFields[field],
						NewType: fieldType,
					})
				}
			}
			for field := range mappedOldFields {
				if _, found = mappedNewFields[field]; !found {
					removedFields = append(removedFields, field)
				}
			}
			var changedResult *types.TlDifferentResult
			if mappedOldObjects[packageName].Result() != newInterface.Result() {
				changedResult = &types.TlDifferentResult{
					OldType: mappedOldObjects[packageName].Result(),
					NewType: newInterface.Result(),
				}
			}
			if (len(newFields)+len(changedFields)+len(removedFields) > 0) || changedResult != nil {
				diff = append(diff, types.TLObjDifferences{
					Object:        newInterface,
					NewFields:     newFields,
					ChangedFields: changedFields,
					RemovedFields: removedFields,
					ChangedResult: changedResult,
				})
			}
		}
	}

	for packageName, oldInterface := range mappedOldObjects {
		if !slices.Contains(foundPackages, packageName) {
			diff = append(diff, types.TLObjDifferences{
				Object:    oldInterface,
				IsDeleted: true,
			})
		}
	}
	return diff
}

func diffKey(object types.TLInterface, schemeLayer int) string {
	layer := object.GetLayer()
	if layer == schemeLayer {
		layer = 0
	}
	return fmt.Sprintf("%s.%d", object.Package(), layer)
}

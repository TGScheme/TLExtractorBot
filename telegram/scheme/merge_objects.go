package scheme

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme/types"
	"TLExtractor/utils"
	"slices"
)

func mergeObjects[T types.TLInterface](old, new []T, isSameLayer bool, patchOs types.PatchOS, remoteOrder bool) []T {
	var orderObjects []string
	objects := make(map[string]T)
	correctNames := make(map[string]string)
	originalObjects := make(map[string]string)
	if environment.LocalStorage.PatchedObjects == nil {
		environment.LocalStorage.PatchedObjects = make(map[types.PatchOS]map[string]*types.PatchInfo)
	}
	if _, ok := environment.LocalStorage.PatchedObjects[patchOs]; !ok {
		environment.LocalStorage.PatchedObjects[patchOs] = make(map[string]*types.PatchInfo)
	}
	for _, oldInterface := range old {
		constructor := ParseConstructor(oldInterface.Constructor())
		objects[constructor] = oldInterface
		originalObjects[oldInterface.Package()] = constructor
		if remoteOrder {
			orderObjects = append(orderObjects, constructor)
		}
	}
	for _, newInterface := range new {
		constructor := ParseConstructor(newInterface.Constructor())
		oldInterface, foundConstructor := objects[constructor]
		if reverseConstructor := originalObjects[newInterface.Package()]; !foundConstructor && len(reverseConstructor) > 0 {
			foundConstructor = true
			constructor = reverseConstructor
			oldInterface = objects[constructor]
		}
		if foundConstructor {
			if oldInterface.Package() != newInterface.Package() {
				correctNames[newInterface.Package()] = oldInterface.Package()
			}
			if patchInfo, ok := environment.LocalStorage.PatchedObjects[patchOs][newInterface.Package()]; ok {
				if patchInfo.PatchedConstructor != oldInterface.Constructor() {
					delete(environment.LocalStorage.PatchedObjects[patchOs], newInterface.Package())
				} else if patchInfo.OldConstructor == newInterface.Constructor() {
					continue
				}
			} else if oldInterface.Constructor() != newInterface.Constructor() && isSameLayer {
				environment.LocalStorage.PatchedObjects[patchOs][newInterface.Package()] = &types.PatchInfo{
					OldConstructor:     newInterface.Constructor(),
					PatchedConstructor: oldInterface.Constructor(),
				}
				continue
			}
			objects[constructor].SetParameters(
				utils.MergeParameters(
					oldInterface.Parameters(),
					newInterface.Parameters(),
					oldInterface.Constructor() == newInterface.Constructor(),
				),
			)
			objects[constructor].SetConstructor(newInterface.Constructor())
			if !remoteOrder {
				orderObjects = append(orderObjects, constructor)
			}
		} else {
			objects[constructor] = newInterface
			orderObjects = append(orderObjects, constructor)
		}
	}
	if remoteOrder {
		slices.Sort(orderObjects[len(old):])
	}
	for _, constructor := range objects {
		if newName, ok := correctNames[constructor.Result()]; ok {
			constructor.SetResult(newName)
		}
		for _, parameter := range constructor.Parameters() {
			if newName, ok := correctNames[parameter.Type]; ok {
				parameter.Type = newName
			}
		}
	}
	var orderedObjects []T
	for _, constructor := range orderObjects {
		orderedObjects = append(orderedObjects, objects[constructor])
	}
	return orderedObjects
}

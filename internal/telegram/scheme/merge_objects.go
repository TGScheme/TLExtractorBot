package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"github.com/TGScheme/TLExtractorBot/internal/utils"
	"slices"
	"strings"
)

func resultLeaf(s string) string {
	if i := strings.LastIndex(s, "."); i >= 0 {
		s = s[i+1:]
	}
	return s
}

func normalizeResult(s string) string {
	return strings.ToLower(resultLeaf(s))
}

func isAbstractResult(s string) bool {
	leaf := resultLeaf(s)
	return leaf != "" && leaf[0] >= 'A' && leaf[0] <= 'Z'
}

func mergeObjects[T types.TLInterface](old, new []T, isSameLayer bool, patches *patchSet, remoteOrder bool, removed map[string]bool) []T {
	var orderObjects []string
	objects := make(map[string]T)
	correctNames := make(map[string]string)
	originalObjects := make(map[string]string)
	for _, oldInterface := range old {
		constructor := ParseConstructor(oldInterface.Constructor())
		objects[constructor] = oldInterface
		originalObjects[oldInterface.Package()] = constructor
		if remoteOrder || oldInterface.IsSecret() {
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
			if patchInfo, ok := patches.get(newInterface.Package()); ok {
				if patchInfo.PatchedConstructor != oldInterface.Constructor() {
					patches.delete(newInterface.Package())
				} else if patchInfo.OldConstructor == newInterface.Constructor() {
					continue
				}
			} else if oldInterface.Constructor() != newInterface.Constructor() && isSameLayer {
				patches.set(newInterface.Package(), &types.PatchInfo{
					OldConstructor:     newInterface.Constructor(),
					PatchedConstructor: oldInterface.Constructor(),
				})
				continue
			} else if oldInterface.Constructor() != newInterface.Constructor() &&
				removed[ParseConstructor(newInterface.Constructor())] {
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

			if newResult := newInterface.Result(); newResult != "" && newInterface.IsMethod() &&
				isAbstractResult(newResult) && normalizeResult(newResult) != "updates" &&
				normalizeResult(newResult) != normalizeResult(oldInterface.Result()) {
				objects[constructor].SetResult(newResult)
			}
			if !remoteOrder && !slices.Contains(orderObjects, constructor) {
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
		params := constructor.Parameters()
		for i := range params {
			if newName, ok := correctNames[params[i].Type]; ok {
				params[i].Type = newName
			}
		}
		constructor.SetParameters(params)
	}
	var orderedObjects []T
	for _, constructor := range orderObjects {
		orderedObjects = append(orderedObjects, objects[constructor])
	}
	return orderedObjects
}

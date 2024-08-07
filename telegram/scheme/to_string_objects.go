package scheme

import (
	"TLExtractor/telegram/scheme/types"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func toStringObjects(objects []types.TLInterface, schemeLayer int) string {
	var tl string
	var keyOrder []string
	var layerOrder []int
	layeredObjects := make(map[int][]types.TLInterface)
	objectsOrder := make(map[string][]types.TLInterface)
	bestLayer := make(map[string]int)
	for _, object := range objects {
		currentLayer := object.GetLayer()
		if currentLayer == 0 {
			currentLayer = schemeLayer
		}
		layeredObjects[currentLayer] = append(layeredObjects[currentLayer], object)
		if !slices.Contains(layerOrder, currentLayer) {
			layerOrder = append(layerOrder, currentLayer)
		}
	}
	slices.Sort(layerOrder)
	for _, layer := range layerOrder {
		objs := layeredObjects[layer]
		if len(layerOrder) > 1 {
			keyOrder = append(keyOrder, fmt.Sprintf("layer %d", layer))
		}
		for _, object := range objs {
			var packageName string
			bestLayer[object.Package()] = layer
			if object.IsMethod() {
				packageName = strings.Split(object.Package(), ".")[0]
			} else {
				packageName = object.Result()
			}
			if object.Result() == "X" {
				packageName = "X"
			}
			packageName += fmt.Sprintf(".%d", layer)
			if !slices.Contains(keyOrder, packageName) {
				keyOrder = append(keyOrder, packageName)
			}
			objectsOrder[packageName] = append(objectsOrder[packageName], object)
		}
	}
	for _, packageName := range keyOrder {
		if strings.HasPrefix(packageName, "layer") {
			tl += fmt.Sprintf("// %s\n", packageName)
			continue
		}
		c := objectsOrder[packageName]
		for _, constructor := range c {
			name := constructor.Package()
			objectLayer := constructor.GetLayer()
			if objectLayer == 0 {
				objectLayer = schemeLayer
			}
			if bestLayer[name] != 0 && bestLayer[name] != objectLayer {
				name += strconv.Itoa(objectLayer)
			}
			tl += fmt.Sprintf("%s#%s", name, ParseConstructor(constructor.Constructor()))
			tempMagicCheck := strings.Split(constructor.Result(), " ")
			magicCheck := tempMagicCheck[len(tempMagicCheck)-1]
			if magicCheck == "X" || magicCheck == "t" {
				tl += fmt.Sprintf(" {%s:Type}", magicCheck)
			}
			if magicCheck == "t" {
				tl += fmt.Sprintf(" # [ %s ]", magicCheck)
			}
			for _, param := range constructor.Parameters() {
				tl += fmt.Sprintf(" %s:%s", param.Name, param.Type)
			}
			tl += " = "
			tl += constructor.Result()
			tl += ";\n"
		}
		tl += "\n"
	}
	return strings.TrimSpace(tl)
}

package scheme

import (
	"strconv"

	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

const maxRestoredNames = 10

func restoreParameterNames(object types.TLInterface, previous []types.Parameter) {
	declared, err := strconv.ParseUint(ParseConstructor(object.Constructor()), 16, 32)
	if err != nil {
		return
	}
	params := object.Parameters()
	matches := func(candidate []types.Parameter) bool {
		return inferIDFromText(objectRepresentation(object.Package(), object.Result(), candidate)) == uint32(declared)
	}
	if matches(params) {
		return
	}
	countTypes := func(list []types.Parameter) map[string]int {
		counts := make(map[string]int, len(list))
		for _, param := range list {
			counts[param.Type]++
		}
		return counts
	}
	oldTypes, newTypes := countTypes(previous), countTypes(params)
	present := make(map[string]bool, len(params))
	for _, param := range params {
		present[param.Name] = true
	}
	var indexes []int
	var names []string
	for i, param := range params {
		if oldTypes[param.Type] != 1 || newTypes[param.Type] != 1 {
			continue
		}
		for _, old := range previous {
			if old.Type == param.Type && old.Name != param.Name && !present[old.Name] {
				indexes = append(indexes, i)
				names = append(names, old.Name)
			}
		}
	}
	if len(indexes) == 0 || len(indexes) > maxRestoredNames {
		return
	}
	for mask := 1; mask < 1<<len(indexes); mask++ {
		candidate := append([]types.Parameter(nil), params...)
		for bit, index := range indexes {
			if mask&(1<<bit) != 0 {
				candidate[index].Name = names[bit]
			}
		}
		if matches(candidate) {
			object.SetParameters(candidate)
			return
		}
	}
}

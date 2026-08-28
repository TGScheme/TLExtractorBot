package scheme

import (
	"regexp"
	"strconv"

	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

var jadxFieldAliasRgx = regexp.MustCompile(`^f\d+([a-zA-Z_][a-zA-Z0-9_]*)$`)

func FixFieldNames(scheme *types.TLFullScheme) int {
	fixed := 0
	for _, part := range []types.TLScheme{scheme.MainApi, scheme.E2EApi} {
		var objects []types.TLInterface
		for _, object := range part.GetConstructors() {
			objects = append(objects, object)
		}
		for _, object := range part.GetMethods() {
			objects = append(objects, object)
		}
		for _, object := range objects {
			fixed += fixFieldNames(object)
		}
	}
	return fixed
}

func fixFieldNames(object types.TLInterface) int {
	declared, err := strconv.ParseUint(ParseConstructor(object.Constructor()), 16, 32)
	if err != nil {
		return 0
	}
	params := object.Parameters()
	renamed := make([]types.Parameter, len(params))
	count := 0
	for i, param := range params {
		renamed[i] = param
		if match := jadxFieldAliasRgx.FindStringSubmatch(param.Name); match != nil {
			renamed[i].Name = match[1]
			count++
		}
	}
	if count == 0 {
		return 0
	}
	name, result := object.Package(), object.Result()
	if inferIDFromText(objectRepresentation(name, result, params)) == uint32(declared) {
		return 0
	}
	if inferIDFromText(objectRepresentation(name, result, renamed)) != uint32(declared) {
		return 0
	}
	object.SetParameters(renamed)
	return count
}

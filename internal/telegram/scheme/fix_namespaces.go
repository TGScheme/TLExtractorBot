package scheme

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

func FixNamespaces(scheme *types.TLFullScheme) int {
	fixed := 0
	for _, part := range []types.TLScheme{scheme.MainApi, scheme.E2EApi} {
		taken := make(map[string]bool)
		var objects []types.TLInterface
		for _, object := range part.GetConstructors() {
			taken[object.Package()] = true
			objects = append(objects, object)
		}
		for _, object := range part.GetMethods() {
			taken[object.Package()] = true
			objects = append(objects, object)
		}
		for _, object := range objects {
			if fixNamespace(object, taken) {
				fixed++
			}
		}
	}
	return fixed
}

func fixNamespace(object types.TLInterface, taken map[string]bool) bool {
	declared, err := strconv.ParseUint(ParseConstructor(object.Constructor()), 16, 32)
	if err != nil {
		return false
	}
	name, result := object.Package(), object.Result()
	if strings.ContainsAny(result, "<> ") {
		return false
	}
	if inferIDFromText(objectRepresentation(name, result, object.Parameters())) == uint32(declared) {
		return false
	}
	for _, variant := range namespaceVariants(name, result) {
		if variant[0] != name && taken[variant[0]] {
			continue
		}
		if inferIDFromText(objectRepresentation(variant[0], variant[1], object.Parameters())) != uint32(declared) {
			continue
		}
		if variant[0] != name {
			delete(taken, name)
			taken[variant[0]] = true
			switch typed := object.(type) {
			case *types.TLConstructor:
				typed.Predicate = variant[0]
			case *types.TLMethod:
				typed.Method = variant[0]
			}
		}
		object.SetResult(variant[1])
		return true
	}
	return false
}

func namespaceVariants(name, result string) [][2]string {
	nameSpace, bareName := splitNamespace(name)
	resultSpace, bareResult := splitNamespace(result)
	variants := [][2]string{
		{bareName, result},
		{name, bareResult},
		{bareName, bareResult},
	}
	for _, namespace := range []string{nameSpace, resultSpace} {
		if namespace == "" {
			continue
		}
		variants = append(variants,
			[2]string{bareName, namespace + "." + bareResult},
			[2]string{namespace + "." + bareName, bareResult},
			[2]string{namespace + "." + bareName, namespace + "." + bareResult},
		)
	}
	return variants
}

func splitNamespace(value string) (string, string) {
	if dot := strings.LastIndex(value, "."); dot >= 0 {
		return value[:dot], value[dot+1:]
	}
	return "", value
}

func objectRepresentation(name, result string, params []types.Parameter) string {
	representation := name
	magic := result
	if fields := strings.Split(result, " "); len(fields) > 1 {
		magic = fields[len(fields)-1]
	}
	if magic == "X" || magic == "t" {
		representation += fmt.Sprintf(" {%s:Type}", magic)
	}
	if magic == "t" {
		representation += fmt.Sprintf(" # [ %s ]", magic)
	}
	for _, param := range params {
		representation += fmt.Sprintf(" %s:%s", param.Name, param.Type)
	}
	return representation + " = " + result
}

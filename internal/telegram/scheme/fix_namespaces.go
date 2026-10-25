package scheme

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

const maxStrippedSuffixes = 4

type namespaceContext struct {
	taken      map[string]bool
	known      map[string]bool
	unique     map[string]string
	namespaces []string
	previous   map[string][][]types.Parameter
}

func FixNamespaces(scheme *types.TLFullScheme, previous ...*types.TLFullScheme) int {
	fixed := 0
	for i, part := range []types.TLScheme{scheme.MainApi, scheme.E2EApi} {
		var objects []types.TLInterface
		objects = append(objects, part.GetConstructors()...)
		objects = append(objects, part.GetMethods()...)
		ctx := namespaceContext{taken: make(map[string]bool), previous: make(map[string][][]types.Parameter)}
		for _, object := range objects {
			ctx.taken[object.Package()] = true
		}
		for _, source := range previous {
			if source == nil {
				continue
			}
			previousPart := source.MainApi
			if i == 1 {
				previousPart = source.E2EApi
			}
			for _, object := range append(previousPart.GetConstructors(), previousPart.GetMethods()...) {
				ctx.previous[object.Package()] = append(ctx.previous[object.Package()], object.Parameters())
			}
		}
		for {
			ctx.known, ctx.unique = qualifiedTypes(objects)
			ctx.namespaces = schemeNamespaces(objects)
			progress := 0
			for _, object := range objects {
				if fixNamespace(object, &ctx) {
					progress++
				}
			}
			if progress == 0 {
				break
			}
			fixed += progress
		}
	}
	return fixed
}

func qualifiedTypes(objects []types.TLInterface) (map[string]bool, map[string]string) {
	known := make(map[string]bool)
	bare := make(map[string]bool)
	for _, object := range objects {
		result := object.Result()
		if strings.ContainsAny(result, "<> ") {
			continue
		}
		if strings.Contains(result, ".") {
			known[result] = true
		} else {
			bare[result] = true
		}
	}
	unique := make(map[string]string)
	ambiguous := make(map[string]bool)
	for result := range known {
		_, name := splitNamespace(result)
		if bare[name] || ambiguous[name] {
			continue
		}
		if _, found := unique[name]; found {
			delete(unique, name)
			ambiguous[name] = true
			continue
		}
		unique[name] = result
	}
	return known, unique
}

func schemeNamespaces(objects []types.TLInterface) []string {
	seen := make(map[string]bool)
	var namespaces []string
	for _, object := range objects {
		if namespace, _ := splitNamespace(object.Package()); namespace != "" && !seen[namespace] {
			seen[namespace] = true
			namespaces = append(namespaces, namespace)
		}
	}
	return namespaces
}

func fixNamespace(object types.TLInterface, ctx *namespaceContext) bool {
	declared, err := strconv.ParseUint(ParseConstructor(object.Constructor()), 16, 32)
	if err != nil {
		return false
	}
	name, result, params := object.Package(), object.Result(), object.Parameters()
	if strings.ContainsAny(result, "<> ") {
		return false
	}
	if inferIDFromText(objectRepresentation(name, result, params)) == uint32(declared) {
		return false
	}
	fileNamespace := ""
	if source, ok := object.(interface{ FileNamespace() string }); ok {
		fileNamespace = source.FileNamespace()
	}
	previous := ctx.previous[name]
	variants := append([]namespaceVariant{{name, result, ""}}, namespaceVariants(name, result, fileNamespace, ctx.namespaces)...)
	for _, variant := range variants {
		if variant.name != name && ctx.taken[variant.name] {
			continue
		}
		candidate, found := matchParams(variant, uint32(declared), params, previous, ctx)
		if !found {
			continue
		}
		if variant.name != name {
			delete(ctx.taken, name)
			ctx.taken[variant.name] = true
			switch typed := object.(type) {
			case *types.TLConstructor:
				typed.Predicate = variant.name
			case *types.TLMethod:
				typed.Method = variant.name
			}
		}
		object.SetResult(variant.result)
		object.SetParameters(candidate)
		return true
	}
	return false
}

type namespaceVariant struct {
	name      string
	result    string
	namespace string
}

func matchParams(variant namespaceVariant, declared uint32, params []types.Parameter, previous [][]types.Parameter, ctx *namespaceContext) ([]types.Parameter, bool) {
	for _, base := range strippedSuffixes(params) {
		candidates := [][]types.Parameter{base}
		if variant.namespace != "" {
			candidates = append(candidates, qualifyParams(base, variant.namespace, ctx.known))
		}
		candidates = append(candidates, qualifyUnique(base, ctx.unique))
		for _, candidate := range candidates {
			if inferIDFromText(objectRepresentation(variant.name, variant.result, candidate)) == declared {
				return candidate, true
			}
			for _, source := range previous {
				if restored, ok := restoreNames(variant.name, variant.result, declared, candidate, source); ok {
					return restored, true
				}
			}
		}
	}
	return nil, false
}

func strippedSuffixes(params []types.Parameter) [][]types.Parameter {
	var indexes []int
	for i, param := range params {
		if strings.HasSuffix(param.Name, "_item") {
			indexes = append(indexes, i)
		}
	}
	variants := [][]types.Parameter{params}
	if len(indexes) == 0 || len(indexes) > maxStrippedSuffixes {
		return variants
	}
	for mask := 1; mask < 1<<len(indexes); mask++ {
		candidate := append([]types.Parameter(nil), params...)
		for bit, index := range indexes {
			if mask&(1<<bit) != 0 {
				candidate[index].Name = strings.TrimSuffix(candidate[index].Name, "_item")
			}
		}
		variants = append(variants, candidate)
	}
	return variants
}

func qualifyUnique(params []types.Parameter, unique map[string]string) []types.Parameter {
	qualified := make([]types.Parameter, len(params))
	for i, param := range params {
		qualified[i] = types.Parameter{Name: param.Name, Type: mapInnerType(param.Type, func(inner string) string {
			if replacement, ok := unique[inner]; ok {
				return replacement
			}
			return inner
		})}
	}
	return qualified
}

func namespaceVariants(name, result, fileNamespace string, others []string) []namespaceVariant {
	nameSpace, bareName := splitNamespace(name)
	resultSpace, bareResult := splitNamespace(result)
	variants := []namespaceVariant{
		{bareName, result, ""},
		{name, bareResult, ""},
		{bareName, bareResult, ""},
	}
	seen := make(map[string]bool)
	for _, namespace := range append([]string{nameSpace, resultSpace, fileNamespace}, others...) {
		if namespace == "" || seen[namespace] {
			continue
		}
		seen[namespace] = true
		names := []string{name, bareName, namespace + "." + bareName}
		if stripped, ok := stripNamespacePrefix(bareName, namespace); ok {
			names = append(names, namespace+"."+stripped)
		}
		results := []string{result, bareResult, namespace + "." + bareResult}
		if stripped, ok := stripNamespacePrefix(bareResult, namespace); ok {
			results = append(results, namespace+"."+stripped)
		}
		for _, candidateName := range names {
			for _, candidateResult := range results {
				variants = append(variants,
					namespaceVariant{candidateName, candidateResult, ""},
					namespaceVariant{candidateName, candidateResult, namespace},
				)
			}
		}
	}
	return variants
}

func stripNamespacePrefix(value, namespace string) (string, bool) {
	if len(value) <= len(namespace) || !strings.EqualFold(value[:len(namespace)], namespace) {
		return "", false
	}
	rest := value[len(namespace):]
	if rest[0] < 'A' || rest[0] > 'Z' {
		return "", false
	}
	if value[0] >= 'a' && value[0] <= 'z' {
		return strings.ToLower(rest[:1]) + rest[1:], true
	}
	return rest, true
}

func qualifyParams(params []types.Parameter, namespace string, known map[string]bool) []types.Parameter {
	qualified := make([]types.Parameter, len(params))
	for i, param := range params {
		qualified[i] = types.Parameter{Name: param.Name, Type: qualifyType(param.Type, namespace, known)}
	}
	return qualified
}

func qualifyType(paramType, namespace string, known map[string]bool) string {
	return mapInnerType(paramType, func(inner string) string {
		if inner == "" || strings.Contains(inner, ".") {
			return inner
		}
		if known[namespace+"."+inner] {
			return namespace + "." + inner
		}
		if stripped, ok := stripNamespacePrefix(inner, namespace); ok && known[namespace+"."+stripped] {
			return namespace + "." + stripped
		}
		if bare := strings.ToUpper(inner[:1]) + inner[1:]; bare != inner && known[namespace+"."+bare] {
			return namespace + "." + inner
		}
		return inner
	})
}

func mapInnerType(paramType string, mapping func(string) string) string {
	prefix, inner := "", paramType
	if before, after, found := strings.Cut(paramType, "?"); found {
		prefix, inner = before+"?", after
	}
	opening := ""
	for strings.HasPrefix(inner, "Vector<") && strings.HasSuffix(inner, ">") {
		opening += "Vector<"
		inner = inner[len("Vector<") : len(inner)-1]
	}
	return prefix + opening + mapping(inner) + strings.Repeat(">", strings.Count(opening, "<"))
}

func splitNamespace(value string) (string, string) {
	if before, after, found := strings.CutLast(value, "."); found {
		return before, after
	}
	return "", value
}

func objectRepresentation(name, result string, params []types.Parameter) string {
	var representation strings.Builder
	representation.WriteString(name)
	magic := result
	if fields := strings.Split(result, " "); len(fields) > 1 {
		magic = fields[len(fields)-1]
	}
	if magic == "X" || magic == "t" {
		representation.WriteString(fmt.Sprintf(" {%s:Type}", magic))
	}
	if magic == "t" {
		representation.WriteString(fmt.Sprintf(" # [ %s ]", magic))
	}
	for _, param := range params {
		representation.WriteString(fmt.Sprintf(" %s:%s", param.Name, param.Type))
	}
	return representation.String() + " = " + result
}

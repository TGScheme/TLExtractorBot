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
		objects = append(objects, part.GetConstructors()...)
		objects = append(objects, part.GetMethods()...)
		for _, object := range objects {
			taken[object.Package()] = true
		}
		for {
			known := qualifiedTypes(objects)
			progress := 0
			for _, object := range objects {
				if fixNamespace(object, taken, known) {
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

func qualifiedTypes(objects []types.TLInterface) map[string]bool {
	known := make(map[string]bool)
	for _, object := range objects {
		if result := object.Result(); strings.Contains(result, ".") && !strings.ContainsAny(result, "<> ") {
			known[result] = true
		}
	}
	return known
}

func fixNamespace(object types.TLInterface, taken, known map[string]bool) bool {
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
	for _, variant := range namespaceVariants(name, result, fileNamespace) {
		if variant.name != name && taken[variant.name] {
			continue
		}
		candidate := params
		if variant.namespace != "" {
			candidate = qualifyParams(params, variant.namespace, known)
		}
		if inferIDFromText(objectRepresentation(variant.name, variant.result, candidate)) != uint32(declared) {
			continue
		}
		if variant.name != name {
			delete(taken, name)
			taken[variant.name] = true
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

func namespaceVariants(name, result, fileNamespace string) []namespaceVariant {
	nameSpace, bareName := splitNamespace(name)
	resultSpace, bareResult := splitNamespace(result)
	variants := []namespaceVariant{
		{bareName, result, ""},
		{name, bareResult, ""},
		{bareName, bareResult, ""},
	}
	seen := make(map[string]bool)
	for _, namespace := range []string{nameSpace, resultSpace, fileNamespace} {
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
	prefix, inner := "", paramType
	if before, after, found := strings.Cut(paramType, "?"); found {
		prefix, inner = before+"?", after
	}
	opening := ""
	for strings.HasPrefix(inner, "Vector<") && strings.HasSuffix(inner, ">") {
		opening += "Vector<"
		inner = inner[len("Vector<") : len(inner)-1]
	}
	if !strings.Contains(inner, ".") && known[namespace+"."+inner] {
		inner = namespace + "." + inner
	}
	return prefix + opening + inner + strings.Repeat(">", strings.Count(opening, "<"))
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

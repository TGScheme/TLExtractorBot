package android

import (
	"TLExtractor/android/ast_parser"
	"TLExtractor/consts"
	"TLExtractor/java"
	"TLExtractor/telegram/scheme/types"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func isOldLayer(className string) bool {
	for _, rgx := range consts.OldLayers {
		if rgx.MatchString(className) {
			return true
		}
	}
	return false
}

var trailingDigitsRgx = regexp.MustCompile(`[0-9]+$`)

func isOldDigitVersion(className string, allNames map[string]bool) bool {
	if !consts.DigitVersionRgx.MatchString(className) {
		return false
	}
	base := trailingDigitsRgx.ReplaceAllString(className, "")
	return allNames[base]
}

func isPlaceholderID(value string) bool {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	id := uint32(n)
	b0, b1, b2 := id>>24, (id>>16)&0xff, (id>>8)&0xff
	return b0 == b1 && b1 == b2
}

func lowerLeaf(name string) string {
	dot := strings.LastIndex(name, ".")
	leaf := name[dot+1:]
	if leaf == "" {
		return name
	}
	return name[:dot+1] + strings.ToLower(leaf[0:1]) + leaf[1:]
}

func extractRawScheme() (*types.RawTLScheme, error) {
	classes, language, err := java.GetClasses()
	if err != nil {
		return nil, err
	}

	layer, err := extractLayer()
	if err != nil {
		return nil, err
	}
	var scheme types.RawTLScheme
	scheme.Layer = layer

	allNames := make(map[string]bool)
	for _, parentClasses := range classes {
		for className := range parentClasses {
			allNames[className] = true
		}
	}

	for _, parent := range slices.Sorted(maps.Keys(classes)) {
		parentClasses := classes[parent]
		for _, className := range slices.Sorted(maps.Keys(parentClasses)) {
			astClass := parentClasses[className]
			bareType := false
			if astClass.ExtendsName == "TLObject" {
				_, oldMethod := astClass.Functions["deserializeResponse"]
				_, newMethod := astClass.Functions["deserializeResponseT"]
				_, hasCtor := astClass.Vars["constructor"]
				_, hasSerialize := astClass.Functions["serializeToStream"]
				if !oldMethod && !newMethod {
					if hasCtor && hasSerialize {
						bareType = true
					} else {
						continue
					}
				}
			}
			if isOldLayer(className) || isOldDigitVersion(className, allNames) ||
				isPlaceholderID(astClass.Vars["constructor"].Value) {
				continue
			}

			parsed, err := ast_parser.Parse(astClass, language)
			if err != nil {
				return nil, fmt.Errorf("error in %s: %w", className, err)
			}

			params := parsed.Params
			for i := range params {
				if len(params[i].Name) > 0 {
					params[i].Name = fixParamName(params[i].Name)
				}
			}

			baseScheme := types.TLBase{
				ID:     astClass.Vars["constructor"].Value,
				Layer:  layer,
				Params: params,
			}

			var hint string
			if parsed.IsMethod {
				hint = parent
			}
			mtProtoName, methLayer := ast_parser.FixTypeName(className, hint, false)
			if layer > methLayer && methLayer != 0 {
				continue
			}

			mtProtoName = lowerLeaf(mtProtoName)

			if parsed.IsMethod {
				if parsed.ReturnType == "" {
					continue
				}
				baseScheme.Type = parsed.ReturnType
				scheme.Methods = append(scheme.Methods, &types.TLMethod{
					TLBase: baseScheme,
					Method: mtProtoName,
				})
			} else {
				if bareType {
					baseScheme.Type, _ = ast_parser.FixTypeName(className, "", true)
				} else {
					baseScheme.Type, _ = ast_parser.FixTypeName(astClass.ExtendsName, astClass.ExtendsPackage, false)
				}
				if baseScheme.Type == "" {
					continue
				}
				scheme.Constructors = append(scheme.Constructors, &types.TLConstructor{
					TLBase:    baseScheme,
					Predicate: mtProtoName,
				})
			}
		}
	}
	return &scheme, nil
}

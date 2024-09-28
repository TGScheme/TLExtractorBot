package android

import (
	"TLExtractor/consts"
	"TLExtractor/java"
	"TLExtractor/telegram/scheme/types"
	"errors"
	"slices"
)

func extractRawScheme() (*types.RawTLScheme, error) {
	rawClasses, err := java.GetRawClasses(isLegacyScheme())
	var ignoreClasses []string
	for _, class := range rawClasses {
		if class.ParentClass == "TLObject" {
			continue
		}
		ignoreClasses = append(ignoreClasses, class.ParentClass)
	}
	if err != nil {
		return nil, err
	}
	var scheme types.RawTLScheme
	for _, class := range rawClasses {
		if slices.Contains(ignoreClasses, class.FullName()) && class.ParentClass == "TLObject" {
			continue
		}
		file, err := extractObject(class)
		if errors.Is(err, consts.ConstructorNotFound) || errors.Is(err, consts.OldLayer) {
			continue
		} else if err != nil {
			return nil, err
		}
		switch file.(type) {
		case *types.TLMethod:
			scheme.Methods = append(scheme.Methods, file.(*types.TLMethod))
		case *types.TLConstructor:
			scheme.Constructors = append(scheme.Constructors, file.(*types.TLConstructor))
		}
	}
	scheme.Layer, err = extractLayer()
	if err != nil {
		return nil, err
	}
	return &scheme, nil
}

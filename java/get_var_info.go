package java

import (
	"TLExtractor/java/types"
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

func GetVarInfo(sitter sitter.Node, source []byte) (*types.AstVar, error) {
	varInfo := &types.AstVar{
		ArrayNesting: 0,
	}
	switch sitter.Kind() {
	case "integral_type", "floating_point_type", "boolean_type", "type_identifier":
		parsedKind := sitter.Utf8Text(source)
		switch parsedKind {
		case "boolean", "Boolean":
			varInfo.Type = "Bool"
		case "float":
			varInfo.Type = "double"
		case "short":
			return nil, nil
		case "String":
			varInfo.Type = "string"
		case "NativeByteBuffer":
			varInfo.Type = "bytes"
		default:
			varInfo.Type = parsedKind
		}
	case "scoped_type_identifier":
		varInfo.Type = sitter.Child(sitter.ChildCount() - 1).Utf8Text(source)
		varInfo.Source = sitter.Child(0).Utf8Text(source)
	case "array_type":
		newVarInfo, err := GetVarInfo(*sitter.Child(0), source)
		if newVarInfo == nil || err != nil {
			return nil, err
		}
		newVarInfo.ArrayNesting = int(sitter.Child(1).ChildCount()) >> 1
		if newVarInfo.ArrayNesting >= 1 && newVarInfo.Type == "byte" {
			newVarInfo.Type = "bytes"
			newVarInfo.ArrayNesting -= 1
		}
		varInfo = newVarInfo
	case "generic_type":
		genericKind := sitter.Child(0).Utf8Text(source)
		switch genericKind {
		case "ArrayList", "Iterator":
			newVarInfo, err := GetVarInfo(*sitter.Child(1).Child(1), source)
			if newVarInfo == nil || err != nil {
				return nil, err
			}
			newVarInfo.ArrayNesting += 1
			varInfo = newVarInfo
		default:

			return nil, nil
		}
	default:
		return nil, fmt.Errorf("unknown type: %s", sitter.Kind())
	}
	return varInfo, nil
}

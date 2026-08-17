package astparser

import (
	javaTypes "github.com/TGScheme/TLExtractorBot/internal/java/types"
	"strconv"
	"strings"
)

const minLayerSuffix = 100

func astTypeToScheme(astType *javaTypes.AstVar) string {
	parsedType := astType.Type
	switch astType.Type {
	case "int", "long", "double", "string", "bytes", "Bool", "true":
	case "Integer":
		parsedType = "int"
	case "Long":
		parsedType = "long"
	case "Double":
		parsedType = "double"
	case "String":
		parsedType = "string"
	case "byte[]", "ByteBuffer":
		parsedType = "bytes"
	case "boolean", "Boolean":
		parsedType = "Bool"
	default:
		t := astType.Type
		if strings.HasPrefix(t, "Tl_") {
			t = "TL_" + strings.TrimPrefix(t, "Tl_")
		}
		if strings.HasPrefix(t, "TL_") {
			parsedType, _ = FixTypeName(t, "", true)
		} else if strings.Contains(t, "_") {
			parsedType, _ = FixTypeName(t, "", true)
		} else if len(astType.Type) > 0 {
			parsedType = strings.ToUpper(astType.Type[0:1]) + astType.Type[1:]
		}
	}
	return strings.Repeat("Vector<", astType.ArrayNesting) + parsedType + strings.Repeat(">", astType.ArrayNesting)
}

func parseStreamPrimitive(javaType string) (string, bool) {
	switch javaType {
	case "readInt32", "writeInt32", "TLDeserializeInt", "deserializeInt":
		return "int", true
	case "readInt64", "writeInt64", "TLDeserializeLong", "deserializeLong":
		return "long", true
	case "readString", "writeString", "deserializeString":
		return "string", true
	case "readDouble", "writeDouble":
		return "double", true
	case "readBoolean", "readBool", "writeBoolean", "writeBool":
		return "Bool", true
	case "readByteBuffer", "readByteArray", "readBytes", "writeByteBuffer", "writeByteArray", "writeBytes":
		return "bytes", true
	}
	return "", false
}

func FixTypeName(name, hint string, allowCasing bool) (string, int) {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		hint = parts[0]
		name = parts[1]
	}
	if name == "TLObject" {
		return "", 0
	}
	name = strings.ReplaceAll(name, "TL_", "")
	if name == "#" {
		return name, 0
	}
	typeInfo := strings.Split(name, "_")

	var packageName string
	if len(hint) > 0 {
		switch hint {
		case "TLRPC", "TLObject":
			packageName = ""
		default:

			packageName = strings.TrimPrefix(hint, "TL_")
		}
	}

	var layer int
	layerIdx := -1
	for i, part := range typeInfo {
		if strings.HasPrefix(part, "layer") {
			suffix := strings.TrimPrefix(part, "layer")
			if suffix == "" && i+1 < len(typeInfo) {
				r, err := strconv.Atoi(typeInfo[i+1])
				if err == nil {
					layer = r
					layerIdx = i
					break
				}
			} else {
				r, err := strconv.Atoi(suffix)
				if err == nil {
					layer = r
					layerIdx = i
					break
				}
			}
		}
	}
	if layerIdx >= 0 {
		for _, part := range typeInfo[layerIdx:] {
			if strings.HasPrefix(part, "old") {
				layer = 1
				break
			}
		}
		typeInfo = typeInfo[:layerIdx]
	}

	if layer == 0 && len(typeInfo) > 1 {
		if r, err := strconv.Atoi(typeInfo[len(typeInfo)-1]); err == nil && r >= minLayerSuffix {
			layer = r
			typeInfo = typeInfo[:len(typeInfo)-1]
		}
	}

	if len(typeInfo) > 0 && strings.HasPrefix(typeInfo[len(typeInfo)-1], "old") {
		typeInfo = typeInfo[:len(typeInfo)-1]
		layer = 1
	}

	if packageName != "" && len(typeInfo) == 1 {
		el := typeInfo[0]
		if strings.HasPrefix(el, packageName) && len(el) > len(packageName) {
			if r := el[len(packageName)]; r >= 'A' && r <= 'Z' {
				rest := el[len(packageName):]
				typeInfo[0] = strings.ToLower(rest[0:1]) + rest[1:]
			}
		}
	}

	if len(typeInfo) > 1 && len(typeInfo[0]) > 0 {
		typeInfo[0] = strings.ToLower(typeInfo[0][0:1]) + typeInfo[0][1:]
	}
	if allowCasing {
		typeInfo[len(typeInfo)-1] = strings.ToUpper(typeInfo[len(typeInfo)-1][0:1]) + typeInfo[len(typeInfo)-1][1:]
	}
	if len(hint) > 0 && len(typeInfo) == 1 && packageName != "" {
		typeInfo = []string{packageName, typeInfo[0]}
	}
	return strings.Join(typeInfo, "."), layer
}

func fieldVar(astClass *javaTypes.AstClass, name string) (*javaTypes.AstVar, bool) {
	if astClass == nil || name == "" {
		return nil, false
	}
	v, ok := astClass.Vars[name]
	return v, ok
}

func fieldType(astClass *javaTypes.AstClass, name string) (string, bool) {
	v, ok := fieldVar(astClass, name)
	if !ok {
		return "", false
	}
	if strings.Contains(v.Type, "ToBeDeprecated") {
		return "", false
	}
	return astTypeToScheme(v), true
}

package android

import (
	"TLExtractor/consts"
	"TLExtractor/java"
	javaTypes "TLExtractor/java/types"
	"TLExtractor/telegram/scheme/types"
	"TLExtractor/utils"
	"regexp"
	"strings"
)

func extractObject(class *javaTypes.RawClass) (types.TLInterface, error) {
	var tl types.TLBase
	var foundVector, isMethod bool
	var deserializedPos, deserializeNesting, serializePos, deserializePos int
	var methodResult string
	compileResult := regexp.MustCompile(`(return|=) *(.*?)\.TLdeserialize`)
	vectorInfo := regexp.MustCompile(`add\(.*readInt([0-9]+)\(`)
	compileConstructorRaw := regexp.MustCompile(`abstractSerializedData[0-9]?\.writeInt32\(([0-9-]+)\)`)

	for pos, line := range class.Content {
		if res := java.GetVarDeclaration(line); res != nil {
			if res.Name == "constructor" {
				tl.ID = res.Value
			}
		} else if java.CheckMethodDec(line, "deserializeResponse") {
			deserializedPos = pos
			deserializeNesting = line.Nesting
			isMethod = true
		} else if deserializedPos != 0 && pos > deserializedPos {
			if line.Nesting >= deserializeNesting {
				if strings.Contains(line.Line, "new TLRPC$Vector()") {
					foundVector = true
				}
				if matches := compileResult.FindAllStringSubmatch(line.Line, -1); len(matches) > 0 {
					formattedType, err := java.FormatType(matches[0][2], true)
					if err != nil {
						return nil, err
					}
					methodResult = formattedType
				} else if matches = vectorInfo.FindAllStringSubmatch(line.Line, -1); len(matches) > 0 {
					if matches[0][1] == "32" {
						methodResult = "int"
					} else if matches[0][1] == "64" {
						methodResult = "long"
					} else {
						return nil, consts.UnknownType
					}
				}
			} else {
				if foundVector {
					methodResult = "Vector<" + methodResult + ">"
				}
				deserializedPos = 0
			}
		} else if java.CheckMethodDec(line, "readParams") {
			deserializePos = pos
		} else if java.CheckMethodDec(line, "serializeToStream") {
			serializePos = pos
		} else if pos > serializePos && serializePos != -1 && line.Nesting >= 2 {
			if matches := compileConstructorRaw.FindAllStringSubmatch(line.Line, -1); len(matches) > 0 && len(tl.ID) == 0 {
				tl.ID = matches[0][1]
			}
		}
	}
	var deserializedParams, serializedParams []types.Parameter
	if serializePos != 0 {
		params, err := extractParams(class, serializePos)
		if err != nil {
			return nil, err
		}
		serializedParams = params
	}
	if deserializePos != 0 {
		params, err := extractParams(class, deserializePos)
		if err != nil {
			return nil, err
		}
		deserializedParams = params
	}
	if len(serializedParams) > len(deserializedParams) {
		if len(deserializedParams) != 0 {
			tl.Params = utils.MergeParameters(deserializedParams, serializedParams, true)
		} else {
			tl.Params = serializedParams
		}
	} else {
		tl.Params = deserializedParams
	}
	if len(tl.ID) == 0 {
		return nil, consts.ConstructorNotFound
	}
	packageName, err := java.FormatType(class.Package, true)
	if err != nil {
		return nil, err
	}
	if len(packageName) == 0 {
		if isMethod {
			packageName = "messages"
		}
	} else {
		packageName = strings.ToLower(packageName)
	}
	if len(packageName) > 0 {
		packageName = packageName + "."
	}
	name, err := java.FormatType(class.Name, true)
	if err != nil {
		return nil, err
	}
	fixedName := strings.ToLower(name[:1]) + name[1:]
	if isMethod {
		tl.Type = methodResult
		return &types.TLMethod{
			TLBase: tl,
			Method: packageName + fixedName,
		}, nil
	} else {
		var parentClass string
		for _, line := range class.Content {
			if className := java.GetParentClass(line); len(className) > 0 {
				parentClass, err = java.FormatType(className, true)
				if err != nil {
					return nil, err
				}
			}
		}
		if parentClass == "Object" {
			parentClass = packageName + name
		}
		tl.Type = parentClass
		return &types.TLConstructor{
			TLBase:    tl,
			Predicate: packageName + fixedName,
		}, nil
	}
}

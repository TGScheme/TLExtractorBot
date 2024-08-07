package java

import (
	"TLExtractor/consts"
	javaTypes "TLExtractor/java/types"
	"errors"
)

func GetVars(lines []javaTypes.LineInfo) ([]javaTypes.VarInfo, error) {
	var parameters []javaTypes.VarInfo
	for _, line := range lines {
		if res := GetVarDeclaration(line); res != nil {
			fixedType, err := FormatType(res.Type, true)
			if errors.Is(err, consts.OldLayer) {
				continue
			} else if err != nil {
				return nil, err
			}
			parameters = append(parameters, javaTypes.VarInfo{
				Name:  res.Name,
				Type:  fixedType,
				Value: res.Value,
			})
		}
	}
	return parameters, nil
}

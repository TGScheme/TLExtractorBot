package java

import (
	"TLExtractor/consts"
	"TLExtractor/utils"
	"regexp"
	"strings"
)

func FormatType(name string, clearTLName bool) (string, error) {
	compileMethodExtend := regexp.MustCompile(`(?i)^TLMethod<(.*)>`)
	if matches := compileMethodExtend.FindAllStringSubmatch(name, -1); len(matches) > 0 {
		name = matches[0][1]
	}
	compile := regexp.MustCompile(`ArrayList<(.*)>`)
	if matches := compile.FindAllStringSubmatch(name, -1); len(matches) > 0 {
		formatted, err := FormatType(matches[0][1], clearTLName)
		if err != nil {
			return "", err
		}
		return "Vector<" + formatted + ">", nil
	}
	fileName := strings.Split(name, "$")
	name = fileName[len(fileName)-1]
	if clearTLName {
		for _, prefix := range []string{"TL", "Tl", "_"} {
			name = strings.TrimPrefix(name, prefix)
		}
	}
	switch strings.ToLower(name) {
	case "bool", "boolean":
		return "Bool", nil
	case "integer", "int":
		return "int", nil
	case "long":
		return "long", nil
	case "double":
		return "double", nil
	case "string":
		return "string", nil
	case "byte[]", "bytes":
		return "bytes", nil
	}
	compile = regexp.MustCompile(`(.*?)_([^_]*?)$`)
	if matches := compile.FindAllStringSubmatch(name, -1); len(matches) > 0 {
		dataName := matches[0][1:]
		name = dataName[0] + "." + utils.Capitalize(dataName[1])
		for _, removed := range consts.OldLayers {
			if removed.MatchString(name) {
				return "", consts.OldLayer
			}
		}
	} else {
		name = utils.Capitalize(name)
	}
	return name, nil
}

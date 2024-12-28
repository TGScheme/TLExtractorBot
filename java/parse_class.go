package java

import (
	"TLExtractor/consts"
	"TLExtractor/java/types"
	"strings"
)

func ParseClass(name, content string) (*types.RawClass, error) {
	prefixFile := strings.Split(name, "$")[0]
	if !strings.HasPrefix(prefixFile, "TL") {
		return nil, consts.NotTLRPC
	}
	name = strings.ReplaceAll(name, ".java", "")
	name, err := FormatType(name, false)
	if err != nil {
		return nil, err
	}
	var tlName types.RawClass
	tlName.Prefix = prefixFile
	if data := strings.Split(name, "."); len(data) > 1 {
		tlName.Name = data[1]
		tlName.Package = data[0]
	} else {
		tlName.Name = name
	}
	tlName.Content = parseLines(content)
	for _, line := range tlName.Content {
		if className := GetParentClass(line); len(className) > 0 {
			tlName.ParentClass, err = FormatType(className, false)
			if err != nil {
				return nil, err
			}
		}
	}
	return &tlName, nil
}

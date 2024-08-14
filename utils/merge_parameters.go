package utils

import (
	"TLExtractor/telegram/scheme/types"
	"regexp"
	"slices"
	"strings"
)

func MergeParameters(old, new []types.Parameter, isSameConstructor bool) []types.Parameter {
	if len(old) == len(new) && isSameConstructor {
		return old
	}
	var mergedList []types.Parameter
	var keys, addableKeys, availableFlags []string
	flagExtractor := regexp.MustCompile(`(flags[0-9]*)\.[0-9]+\?`)
	i, j := 0, 0
	for _, content := range old {
		keys = append(keys, content.Name)
	}
	for _, content := range new {
		addableKeys = append(addableKeys, content.Name)
		if content.Type == "#" {
			availableFlags = append(availableFlags, content.Name)
		}
	}
	for i < len(old) || j < len(new) {
		if i < len(old) {
			content := old[i]
			res := flagExtractor.FindAllStringSubmatch(content.Type, -1)
			isFlagAddable := len(res) > 0 && slices.Contains(availableFlags, res[0][1]) && isSameConstructor
			if slices.Contains(addableKeys, content.Name) || isFlagAddable {
				mergedList = append(mergedList, content)
			}
			i++
		}
		if j < len(new) {
			content := new[j]
			if !slices.Contains(keys, content.Name) {
				mergedList = append(mergedList, content)
				keys = append(keys, content.Name)
			}
			j++
		}
	}
	if !isSameConstructor {
		for _, content := range new {
			for pos, mergedContent := range mergedList {
				if content.Name == mergedContent.Name {
					if strings.HasSuffix(mergedContent.Type, "Bool") && strings.HasSuffix(content.Type, "true") {
						break
					}
					mergedList[pos].Type = content.Type
					break
				}
			}
		}
	}
	return mergedList
}

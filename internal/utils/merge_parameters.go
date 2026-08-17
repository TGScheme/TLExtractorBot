package utils

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"regexp"
	"slices"
	"strings"
)

func MergeParameters(old, new []types.Parameter, isSameConstructor bool) []types.Parameter {
	if isSameConstructor {
		return SortFlagBools(appendUnknownFlagBools(old, new))
	}
	var mergedList []types.Parameter
	var keys, addableKeys, availableFlags []string
	flagExtractor := regexp.MustCompile(`(flags[0-9]*)\.[0-9]+\?`)
	i, j := 0, 0
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
			if len(res) > 0 && slices.Contains(availableFlags, res[0][1]) &&
				!slices.Contains(addableKeys, content.Name) &&
				!slices.Contains(keys, content.Name) &&
				isSameConstructor {
				mergedList = append(mergedList, content)
				keys = append(keys, content.Name)
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

func appendUnknownFlagBools(old, new []types.Parameter) []types.Parameter {
	known := make(map[string]bool, len(old))
	declaredFlags := make(map[string]bool)
	usedBits := make(map[string]bool)
	for _, p := range old {
		known[p.Name] = true
		if p.Type == "#" {
			declaredFlags[p.Name] = true
		}
		if match := flagUseRgx.FindStringSubmatch(p.Type); match != nil {
			usedBits[match[1]+"."+match[2]] = true
		}
	}
	merged := slices.Clone(old)
	for _, p := range new {
		match := flagBoolRgx.FindStringSubmatch(p.Type)
		if match == nil || known[p.Name] || !declaredFlags[match[1]] {
			continue
		}

		if usedBits[match[1]+"."+match[2]] {
			continue
		}
		known[p.Name] = true
		merged = append(merged, p)
	}
	return merged
}

package android

import (
	"TLExtractor/utils"
	"strings"
)

func fixParamName(name string) string {
	newName := utils.SnakeCase(name)
	if newName != "is_admin" {
		newName = strings.TrimPrefix(newName, "is_")
	}
	if newName != name {
		newName = strings.TrimPrefix(newName, "web_")
		newName = strings.ReplaceAll(newName, "__b", "_B")
	}
	newName = strings.TrimSuffix(newName, "_item")
	if newName == "hash2" {
		newName = "hash"
	}
	newName = strings.TrimPrefix(newName, "_")
	return newName
}

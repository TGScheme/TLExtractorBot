package utils

import (
	"strings"
	"unicode"
)

func SnakeCase(name string) string {
	var nameNew strings.Builder
	nameNew.WriteString(strings.ToLower(name[:1]))
	for _, v := range name[1:] {
		if unicode.IsUpper(v) {
			nameNew.WriteString("_")
		}
		nameNew.WriteString(strings.ToLower(string(v)))
	}
	return nameNew.String()
}

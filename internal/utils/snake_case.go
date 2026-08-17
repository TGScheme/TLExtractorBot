package utils

import (
	"strings"
	"unicode"
)

func SnakeCase(name string) string {
	nameNew := strings.ToLower(name[:1])
	for _, v := range name[1:] {
		if unicode.IsUpper(v) {
			nameNew += "_"
		}
		nameNew += strings.ToLower(string(v))
	}
	return nameNew
}

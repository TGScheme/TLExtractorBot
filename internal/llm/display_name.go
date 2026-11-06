package llm

import (
	"strings"
	"unicode"
)

func DisplayName(model string) string {
	parts := strings.Split(strings.TrimPrefix(model, "models/"), "-")
	if len(parts) == 0 || parts[0] == "" {
		return model
	}
	var words []string
	switch {
	case parts[0] == "gpt" && len(parts) > 1:
		words = append(words, "GPT-"+parts[1])
		parts = parts[2:]
	case len(parts[0]) > 1 && parts[0][0] == 'o' && unicode.IsDigit(rune(parts[0][1])):
		words = append(words, parts[0])
		parts = parts[1:]
	}
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		words = append(words, string(runes))
	}
	return strings.Join(words, " ")
}

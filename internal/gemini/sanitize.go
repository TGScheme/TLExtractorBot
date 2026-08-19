package gemini

import (
	"regexp"
	"strings"
)

var (
	boldMarkup   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicMarkup = regexp.MustCompile(`(^|[^*\w])\*([^\s*](?:[^*\n]*?[^\s*])?)\*([^*\w]|$)`)
	codeMarkup   = regexp.MustCompile("`([^`\n]+?)`")
	escaper      = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
)

func sanitize(text string) string {
	text = escaper.Replace(strings.TrimSpace(text))
	text = codeMarkup.ReplaceAllString(text, "<code>$1</code>")
	text = boldMarkup.ReplaceAllString(text, "<b>$1</b>")
	text = italicMarkup.ReplaceAllString(text, "$1<i>$2</i>$3")
	return text
}

func sanitizeAll(texts []string) []string {
	var out []string
	for _, text := range texts {
		if clean := sanitize(text); clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

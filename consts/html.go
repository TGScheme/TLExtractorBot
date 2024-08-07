package consts

var SupportedHtmlTags = []string{
	"a",
	"aside",
	"b",
	"blockquote",
	"br",
	"code",
	"em",
	"figcaption",
	"figure",
	"h3",
	"h4",
	"hr",
	"i",
	"iframe",
	"img",
	"li",
	"ol",
	"p",
	"pre",
	"s",
	"strong",
	"u",
	"ul",
	"video",
}

var TagRequiredAttrs = map[string][]string{
	"a":      {"href"},
	"iframe": {"src"},
	"img":    {"src"},
	"video":  {"src"},
}

var TagUnsupportedChildren = []string{
	"img",
	"iframe",
	"video",
	"br",
	"hr",
}

var UnclosedTags = []string{
	"br",
	"hr",
	"img",
}

var SpecialChars = map[string]string{
	"amp":  "&",
	"lt":   "<",
	"gt":   ">",
	"quot": "\"",
	"apos": "'",
}

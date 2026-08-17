package scheme

import (
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"regexp"
	"strings"
)

var flagPrefixRe = regexp.MustCompile(`^flags\d*\.\d+\?`)

var tlPrimitives = map[string]bool{
	"true": true, "int": true, "long": true, "double": true,
	"string": true, "bytes": true, "Bool": true, "#": true,
	"int128": true, "int256": true,
}

func canonicalTypeIndex(objs []types.TLInterface) map[string]string {
	seen := make(map[string]map[string]bool)
	add := func(full string) {
		full = strings.TrimSpace(full)
		if full == "" || full == "X" || full == "#" {
			return
		}
		leaf := full
		if i := strings.LastIndex(leaf, "."); i >= 0 {
			leaf = leaf[i+1:]
		}
		ll := strings.ToLower(leaf)
		if ll == "" {
			return
		}
		if seen[ll] == nil {
			seen[ll] = make(map[string]bool)
		}
		seen[ll][full] = true
	}
	for _, o := range objs {
		add(o.Result())
	}
	idx := make(map[string]string, len(seen))
	for leaf, spellings := range seen {
		if len(spellings) == 1 {
			for s := range spellings {
				idx[leaf] = s
			}
		}
	}
	return idx
}

func canonicalPredicateIndex(objs []types.TLInterface) map[string]string {
	lowerLeafOf := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		if i := strings.LastIndex(s, "."); i >= 0 {
			s = s[i+1:]
		}
		return s
	}

	typeLeaves := make(map[string]bool)
	for _, o := range objs {
		typeLeaves[lowerLeafOf(o.Result())] = true
	}
	seen := make(map[string]map[string]bool)
	for _, o := range objs {
		if o.IsMethod() {
			continue
		}
		leaf := lowerLeafOf(o.Package())
		result := strings.TrimSpace(o.Result())
		if leaf == "" || result == "" || typeLeaves[leaf] {
			continue
		}
		if seen[leaf] == nil {
			seen[leaf] = make(map[string]bool)
		}
		seen[leaf][result] = true
	}
	idx := make(map[string]string, len(seen))
	for leaf, results := range seen {
		if len(results) == 1 {
			for r := range results {
				idx[leaf] = r
			}
		}
	}
	return idx
}

func canonicalizeType(t string, idx, predIdx map[string]string) string {
	prefix := flagPrefixRe.FindString(t)
	t = t[len(prefix):]
	var open, closing string
	for strings.HasPrefix(t, "Vector<") && strings.HasSuffix(t, ">") {
		open += "Vector<"
		closing += ">"
		t = t[len("Vector<") : len(t)-1]
	}
	if tlPrimitives[t] {
		return prefix + open + t + closing
	}
	leaf := t
	if i := strings.LastIndex(leaf, "."); i >= 0 {
		leaf = leaf[i+1:]
	}
	if canon, ok := idx[strings.ToLower(leaf)]; ok {
		if canon != t {
			t = canon
		}
	} else if canon, ok := predIdx[strings.ToLower(leaf)]; ok {
		t = canon
	}

	dot := strings.LastIndex(t, ".")
	if l := t[dot+1:]; l != "" && l[0] >= 'a' && l[0] <= 'z' {
		t = t[:dot+1] + strings.ToUpper(l[0:1]) + l[1:]
	}
	return prefix + open + t + closing
}

func canonicalizeScheme[T types.TLInterface](objs []T, idx, predIdx map[string]string) {
	for _, o := range objs {
		o.SetResult(canonicalizeType(o.Result(), idx, predIdx))
		params := o.Parameters()
		for i := range params {
			params[i].Type = canonicalizeType(params[i].Type, idx, predIdx)
		}
		o.SetParameters(params)
	}
}

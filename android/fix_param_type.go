package android

import "TLExtractor/consts"

func fixParamType(name string) string {
	for rgx, repl := range consts.BrokenTypes {
		name = rgx.ReplaceAllString(name, repl)
	}
	return name
}

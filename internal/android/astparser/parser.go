package astparser

import (
	javaTypes "github.com/TGScheme/TLExtractorBot/internal/java/types"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"github.com/TGScheme/TLExtractorBot/internal/utils"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type ParseResult struct {
	Params     []types.Parameter
	ReturnType string
	IsMethod   bool
}

func deserializeResponseFn(astClass *javaTypes.AstClass) (string, bool) {
	if body, ok := astClass.Functions["deserializeResponse"]; ok {
		return body, true
	}
	if body, ok := astClass.Functions["deserializeResponseT"]; ok {
		return body, true
	}
	return "", false
}

func Parse(astClass *javaTypes.AstClass, language *sitter.Language) (ParseResult, error) {
	drBody, isMethod := deserializeResponseFn(astClass)

	var paramSource string
	if isMethod {
		paramSource = astClass.Functions["serializeToStream"]
	} else {
		paramSource = astClass.Functions["readParams"]
	}

	res := ParseResult{IsMethod: isMethod}

	if paramSource != "" {
		block, src, closer := wrapBlock(language, paramSource)
		if block != nil {
			w := newWalker(astClass, src)
			if isMethod {
				w.walkSerializeBlock(block)
			} else {
				w.walkBlock(block, 0)
			}
			res.Params = w.result()
		}
		closer()
	}

	if !isMethod {
		res.Params = mergeWriteSideBools(astClass, language, res.Params)
	}

	res.Params = utils.SortFlagBools(res.Params)

	if isMethod {
		if drBody != "" {
			block, src, closer := wrapBlock(language, drBody)
			if block != nil {
				w := newWalker(astClass, src)
				res.ReturnType = w.returnTypeFromDeserialize(block)
			}
			closer()
		}
	}
	return res, nil
}

func mergeWriteSideBools(astClass *javaTypes.AstClass, language *sitter.Language, params []types.Parameter) []types.Parameter {
	serBody := astClass.Functions["serializeToStream"]
	if serBody == "" {
		return params
	}
	block, src, closer := wrapBlock(language, serBody)
	defer closer()
	if block == nil {
		return params
	}
	w := newWalker(astClass, src)
	w.walkSerializeBlock(block)

	have := make(map[string]bool, len(params))
	for _, p := range params {
		have[p.Name] = true
	}

	var bools []types.Parameter
	for _, p := range w.result() {
		if have[p.Name] {
			continue
		}
		if strings.HasSuffix(p.Type, "?true") && strings.Contains(p.Type, ".") {
			bools = append(bools, p)
			have[p.Name] = true
		}
	}
	if len(bools) == 0 {
		return params
	}

	byField := make(map[string][]types.Parameter)
	for _, b := range bools {
		field := b.Type[:strings.Index(b.Type, ".")]
		byField[field] = append(byField[field], b)
	}
	out := make([]types.Parameter, 0, len(params)+len(bools))
	for _, p := range params {
		out = append(out, p)
		if p.Type == "#" {
			out = append(out, byField[p.Name]...)
		}
	}
	return out
}

func wrapBlock(language *sitter.Language, body string) (*sitter.Node, []byte, func()) {
	inner := body
	if len(inner) >= 2 && inner[0] == '{' && inner[len(inner)-1] == '}' {
		inner = inner[1 : len(inner)-1]
	}
	return parseMethodBlock(language, inner)
}

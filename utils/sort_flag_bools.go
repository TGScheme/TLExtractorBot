package utils

import (
	"TLExtractor/telegram/scheme/types"
	"regexp"
	"slices"
	"strconv"
)

var flagBoolRgx = regexp.MustCompile(`^(flags[0-9]*)\.([0-9]+)\?true$`)

var flagUseRgx = regexp.MustCompile(`^(flags[0-9]*)\.([0-9]+)\?`)

func SortFlagBools(params []types.Parameter) []types.Parameter {
	flagAt := make(map[string]int)
	for i, p := range params {
		if p.Type == "#" {
			flagAt[p.Name] = i
		}
	}
	if len(flagAt) == 0 {
		return params
	}
	type boolParam struct {
		bit   int
		param types.Parameter
	}
	groups := make(map[string][]boolParam)
	rest := make([]types.Parameter, 0, len(params))
	for _, p := range params {
		match := flagBoolRgx.FindStringSubmatch(p.Type)
		if match == nil {
			rest = append(rest, p)
			continue
		}
		if _, ok := flagAt[match[1]]; !ok {
			rest = append(rest, p)
			continue
		}
		bit, _ := strconv.Atoi(match[2])
		groups[match[1]] = append(groups[match[1]], boolParam{bit: bit, param: p})
	}
	if len(groups) == 0 {
		return params
	}
	for _, group := range groups {
		slices.SortStableFunc(group, func(a, b boolParam) int { return a.bit - b.bit })
	}
	out := make([]types.Parameter, 0, len(params))
	for _, p := range rest {
		out = append(out, p)
		if p.Type != "#" {
			continue
		}
		for _, b := range groups[p.Name] {
			out = append(out, b.param)
		}
	}
	return out
}

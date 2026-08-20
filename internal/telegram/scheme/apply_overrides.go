package scheme

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
)

//go:embed overrides.tl
var overridesFile string

var overrides []tlObject

func init() {
	parsed, problems := parseTLText("overrides.tl", overridesFile)
	for _, problem := range problems {
		gologging.Fatal(fmt.Errorf("invalid override: %s", problem))
	}
	for _, override := range parsed {
		if !override.matchesInferredID() {
			gologging.Fatal(fmt.Errorf(
				"the override of %s does not match its own constructor id #%s",
				override.name, override.id,
			))
		}
	}
	overrides = parsed
}

func ApplyOverrides(scheme *types.TLFullScheme) int {
	if len(overrides) == 0 {
		return 0
	}
	index := make(map[uint32]types.TLInterface)
	for _, part := range []types.TLScheme{scheme.MainApi, scheme.E2EApi} {
		for _, object := range append(part.GetConstructors(), part.GetMethods()...) {
			id, err := strconv.ParseUint(ParseConstructor(object.Constructor()), 16, 32)
			if err != nil {
				continue
			}
			index[uint32(id)] = object
		}
	}
	applied := 0
	for _, override := range overrides {
		id, err := strconv.ParseUint(override.id, 16, 32)
		if err != nil {
			continue
		}
		target, found := index[uint32(id)]
		if !found {
			gologging.Warn(fmt.Sprintf("scheme: the override of %s is stale, no object uses #%s", override.name, override.id))
			continue
		}
		switch typed := target.(type) {
		case *types.TLConstructor:
			typed.Predicate = override.name
		case *types.TLMethod:
			typed.Method = override.name
		}
		target.SetParameters(override.params)
		target.SetResult(override.result)
		applied++
	}
	return applied
}

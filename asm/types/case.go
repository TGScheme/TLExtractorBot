package types

import (
	"maps"
	"slices"
)

type Case struct {
	Regions map[int64]int64
}

func (c Case) IsInside(offset int64) bool {
	for region, size := range c.Regions {
		if offset >= region && offset < region+size {
			return true
		}
	}
	return false
}

func (c Case) Offset() int64 {
	return slices.Min(slices.Collect(maps.Keys(c.Regions)))
}

package asm

import (
	"TLExtractor/asm/types"
	"golang.org/x/arch/arm64/arm64asm"
	"maps"
	"slices"
)

func switchCBZOrdered(instructions []arm64asm.Inst, switchInfo types.Switch, caseOffset int) ([]arm64asm.Inst, error) {
	var caseCount int
	var blocks [][]*types.Case
	var lastOrder types.CBZOrderKind
	for _, cbz := range getCBZOrderInfo(instructions, switchInfo) {
		var firstElement, lastElement *types.Case
		var foundFirst bool
		for _, caseInfo := range switchInfo.Cases[caseCount*2 : min(caseCount*2+2, len(switchInfo.Cases))] {
			if caseInfo.Offset() >= cbz.FirstOffset && !foundFirst {
				firstElement = &caseInfo
				foundFirst = true
			} else {
				lastElement = &caseInfo
			}
		}
		lastOrder = cbz.Order
		if cbz.Order == types.ReverseOrder {
			firstElement, lastElement = lastElement, firstElement
		}
		caseCount++
		blocks = append(blocks, []*types.Case{firstElement, lastElement})
	}
	if caseCount*2 < len(switchInfo.Cases) {
		for _, caseInfo := range switchInfo.Cases[caseCount*2:] {
			blocks = append(blocks, []*types.Case{&caseInfo})
		}
	}
	var orderedCases []*types.Case
	for _, block := range blocks {
		orderedCases = append(orderedCases, block...)
	}
	_ = lastOrder
	if lastOrder == types.ReverseOrder && caseCount*2 < len(switchInfo.Cases) {
		slices.Reverse(orderedCases)
	}
	caseInfo := orderedCases[caseOffset]
	var instructionsToReturn []arm64asm.Inst
	orderedRegions := slices.Collect(maps.Keys(caseInfo.Regions))
	slices.Sort(orderedRegions)
	for _, offset := range orderedRegions {
		size := caseInfo.Regions[offset]
		instructionsToReturn = append(instructionsToReturn, instructions[offset:offset+size]...)
	}
	return instructionsToReturn, nil
}

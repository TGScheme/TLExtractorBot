package asm

import (
	"TLExtractor/asm/types"
	"fmt"
	"golang.org/x/arch/arm64/arm64asm"
)

var NoCasesError = fmt.Errorf("no cases found")

func ExtractFromSwitch(instructions []arm64asm.Inst, caseOffset int) ([]arm64asm.Inst, error) {
	switchInfo := getSwitch(instructions)
	if caseOffset >= len(switchInfo.Cases) {
		return nil, fmt.Errorf("case offset %d is out of range", caseOffset)
	}
	if len(switchInfo.Cases) == 0 {
		return nil, NoCasesError
	}
	switch switchInfo.Kind {
	case types.SwitchCBZKind:
		return switchCBZOrdered(instructions, switchInfo, caseOffset)
	case types.SwitchTBNZKind:
		fmt.Println("TBNZ", len(switchInfo.Cases))
		return switchTBNZOrdered(instructions, switchInfo, caseOffset)
	default:
		return nil, fmt.Errorf("unsupported switch type with %d cases", len(switchInfo.Cases))
	}
}

package asm

import (
	"TLExtractor/asm/types"
	"fmt"
	"golang.org/x/arch/arm64/arm64asm"
)

func switchTBNZOrdered(instructions []arm64asm.Inst, switchInfo types.Switch, caseOffset int) ([]arm64asm.Inst, error) {
	//for _, instruction := range instructions {
	//	fmt.Println(instruction)
	//}
	return nil, fmt.Errorf("invalid tbnz instruction")
}

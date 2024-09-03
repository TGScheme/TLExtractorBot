package ios

import (
	"fmt"
	"golang.org/x/arch/arm64/arm64asm"
	"reflect"
)

func extractConstructor(instructions []arm64asm.Inst) (string, error) {
	for i, instruction := range instructions {
		if instruction.Op == arm64asm.MOVK {
			if imm, ok := instructions[i-1].Args[1].(arm64asm.Imm); ok {
				shiftInfo := reflect.ValueOf(instruction.Args[1])
				shift := shiftInfo.FieldByName("shift").Uint()
				lowValue := uint64(imm.Imm)
				highValue := shiftInfo.FieldByName("imm").Uint()
				return fmt.Sprintf("%x", highValue<<shift|lowValue), nil
			}
		}
	}
	return "", fmt.Errorf("constructor not found")
}

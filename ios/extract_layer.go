package ios

import (
	"TLExtractor/asm"
	"errors"
	"github.com/Laky-64/goswift/proxy"
	"golang.org/x/arch/arm64/arm64asm"
)

func extractLayer(file *proxy.Context) (int, error) {
	classes, err := file.GetObjCClasses()
	if err != nil {
		return 0, err
	}
	for _, class := range classes {
		for _, method := range class.InstanceMethods {
			if method.Name == "currentLayer" {
				instructions, err := asm.GetInstructions(file, method.ImpVMAddr)
				if err != nil {
					return 0, err
				}
				return int(instructions[0].Args[1].(arm64asm.Imm).Imm), nil
			}
		}
	}
	return 0, errors.New("layer not found")
}

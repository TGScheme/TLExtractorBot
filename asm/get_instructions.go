package asm

import (
	"github.com/Laky-64/goswift/proxy"
	"golang.org/x/arch/arm64/arm64asm"
)

func GetInstructions(file *proxy.Context, addr uint64) ([]arm64asm.Inst, error) {
	f, err := file.GetFunctionForVMAddr(addr)
	if err != nil {
		return nil, err
	}
	data, err := file.GetFunctionData(f)
	if err != nil {
		return nil, err
	}
	var instructions []arm64asm.Inst
	for i := 0; i < len(data); i += 4 {
		decode, err := arm64asm.Decode(data[i:])
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, decode)
	}
	return instructions, nil
}

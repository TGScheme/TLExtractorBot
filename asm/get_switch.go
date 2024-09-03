package asm

import (
	"TLExtractor/asm/types"
	"fmt"
	"github.com/Laky-64/gologging"
	"golang.org/x/arch/arm64/arm64asm"
)

func getSwitch(instructions []arm64asm.Inst) types.Switch {
	var regCase arm64asm.Reg
	var switchInfo types.Switch
	var foundReg bool
	var foundKind bool
	for i, inst := range instructions {
		//fmt.Println(inst)
		if !foundKind && len(switchInfo.Cases) == 0 {
			switch inst.Op {
			case arm64asm.CBZ:
				foundKind = true
				switchInfo.Kind = types.SwitchCBZKind
			case arm64asm.B:
				if cond, ok := inst.Args[0].(arm64asm.Cond); ok && cond.String() == "EQ" {
					foundKind = true
					switchInfo.Kind = types.SwitchEQKind
				}
			case arm64asm.TBNZ:
				foundKind = true
				switchInfo.Kind = types.SwitchTBNZKind
			default:
				switchInfo.Kind = types.SwitchKindUnknown
			}
		}
		if inst.Op == arm64asm.TBZ && inst.Args[1].(arm64asm.Imm).Imm == 0 {
			reg := inst.Args[0].(arm64asm.Reg)
			if !foundReg {
				foundReg = true
				regCase = reg
			} else if regCase != reg {
				continue
			}
			currRegion := int64(i)
			regionsSize := map[int64]int64{
				currRegion: 0,
			}
			for j := i; j < len(instructions); j++ {
				op := instructions[j].Op
				if op == arm64asm.RET || op == arm64asm.BRK {
					break
				}
				if op == arm64asm.B {
					if pc, ok := instructions[j].Args[0].(arm64asm.PCRel); ok {
						if pc < 0 {
							break
						}
						currRegion = int64(j) + int64(pc)/4
						regionsSize[currRegion] = 0
						j += int(pc) / 4
					}
				}
				regionsSize[currRegion]++
			}
			switchInfo.Cases = append(switchInfo.Cases, types.Case{
				Regions: regionsSize,
			})
		}
	}
	//switchInfo.Kind = types.SwitchTBZKind
	if switchInfo.Kind == types.SwitchEQKind && len(switchInfo.Cases) != 2 {
		gologging.Warn(fmt.Sprintf("EQ switch with %d cases", len(switchInfo.Cases)))
	}
	switchInfo.Reg = regCase
	return switchInfo
}

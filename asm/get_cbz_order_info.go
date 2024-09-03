package asm

import (
	"TLExtractor/asm/types"
	"golang.org/x/arch/arm64/arm64asm"
)

func getCBZOrderInfo(instructions []arm64asm.Inst, switchInfo types.Switch) []types.CBZInfo {
	var cbzList []types.CBZInfo
	for i, inst := range instructions {
		var isInsideSwitch bool
		for _, caseInfo := range switchInfo.Cases {
			if caseInfo.IsInside(int64(i)) {
				isInsideSwitch = true
				break
			}
		}
		var isSwitch bool
		switch inst.Op {
		case arm64asm.B:
			if cond, ok := inst.Args[0].(arm64asm.Cond); ok && cond.String() == "EQ" {
				isSwitch = true
			}
		case arm64asm.CBZ:
			isSwitch = true
		default:
			continue
		}
		if isSwitch && !isInsideSwitch {
			pcSkip := int64(inst.Args[1].(arm64asm.PCRel))/4 + int64(i)
			var isValid bool
			for j := int(pcSkip); j < len(instructions); j++ {
				if instructions[j].Op == arm64asm.B {
					break
				}
				if instructions[j].Op == arm64asm.TBZ {
					reg := instructions[j].Args[0].(arm64asm.Reg)
					if reg == switchInfo.Reg {
						isValid = true
					}
					break
				}
			}
			if !isValid {
				continue
			}
			var info types.CBZInfo
			info.Offset = i
			info.Order = types.NormalOrder
			info.FirstOffset = pcSkip

			var foundMove bool
			for _, subInst := range instructions[i:] {
				if subInst.Op == arm64asm.B || subInst.Op == arm64asm.CBNZ {
					break
				} else if subInst.Op == arm64asm.CMP && foundMove {
					info.Order = types.ReverseOrder
				} else if subInst.Op == arm64asm.MOV {
					foundMove = true
				} else if subInst.Op == arm64asm.TBZ {
					if len(switchInfo.Cases) == 2 {
						info.Order = types.ReverseOrder
					}
					break
				}
			}
			cbzList = append(cbzList, info)
		}
	}
	return cbzList
}

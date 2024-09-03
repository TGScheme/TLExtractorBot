package asm

import (
	"TLExtractor/environment"
	"TLExtractor/telegram/scheme"
	"bytes"
	"fmt"
	"golang.org/x/arch/arm64/arm64asm"
	"os"
	"sort"
	"strings"
)

func findSkips(numbers []int) []int {
	sort.Ints(numbers) // Ordina i numeri
	var skips []int
	for i := 1; i < len(numbers); i++ {
		if numbers[i] != numbers[i-1]+1 {
			for j := numbers[i-1] + 1; j < numbers[i]; j++ {
				skips = append(skips, j)
			}
		}
	}
	return skips
}

func DebugAsm(instructions []arm64asm.Inst, toBeDeleted []string) {
	fmt.Println("Skips:")
	for _, t := range toBeDeleted {
		fmt.Println(t)
	}
	/*
		SCHEME ATTEMPT: (TEST_CASES/debug.asm)
		1
		2
		3
		4
		5
		6
		7
		8
		MISSING 9
		10
		11
		12
		13
		14
		15
		16
		17
		18
		19
		MISSING 20
		21
		22
		MISSING 23
		MISSING 24
		25
		26
		MISSING 27
		28
		9  SHOULD NOT BE HERE
		20 SHOULD NOT BE HERE
		23 SHOULD NOT BE HERE
		24 SHOULD NOT BE HERE
		27 SHOULD NOT BE HERE
		29
	*/
	switchInfo := getSwitch(instructions)
	var outputFileTemp bytes.Buffer
	var clearASMFIle bytes.Buffer
	var switchIdx int
	var wasInside bool
	//prevIdx := 0
	reverseNames := make(map[string]string)
	for _, obj := range environment.LocalStorage.PreviewLayer.MainApi.Constructors {
		reverseNames[scheme.ParseConstructor(obj.Constructor())] = obj.Package()
	}
	var foundConstructor int
	for i, instruction := range instructions {
		var isInside bool
		for idx, caseInfo := range switchInfo.Cases[switchIdx:] {
			off := int(caseInfo.Offset())
			size := int(caseInfo.Regions[caseInfo.Offset()])
			if i >= off && i < off+size {
				switchIdx = idx + switchIdx
				isInside = true
				break
			}
		}
		if isInside && !wasInside {
			c, _ := extractConstructor(instructions[i:])
			foundConstructor++
			reversedName := reverseNames[c]
			var foundPos int
			for x, d := range toBeDeleted {
				if strings.TrimPrefix(d, "Api.") == reversedName {
					foundPos = x
					break
				}
			}
			outputFileTemp.WriteString(fmt.Sprintf("\n\n//CASE %d: %s (%s) original_order=%d\n", switchIdx+1, c, reversedName, foundPos+1))
		}
		outputFileTemp.WriteString(instruction.String())
		if instruction.Op == arm64asm.B || instruction.Op == arm64asm.CBZ || instruction.Op == arm64asm.CBNZ || instruction.Op == arm64asm.TBZ || instruction.Op == arm64asm.TBNZ {
			var idx int
			if _, ok := instruction.Args[0].(arm64asm.Cond); ok || instruction.Op == arm64asm.CBZ || instruction.Op == arm64asm.CBNZ {
				idx = 1
			}
			if instruction.Op == arm64asm.TBZ || instruction.Op == arm64asm.TBNZ {
				idx = 2
			}
			if pc, ok := instruction.Args[idx].(arm64asm.PCRel); ok {
				newPos := i + int(pc/4)
				outputFileTemp.WriteString(fmt.Sprintf(" // PCRel %d", newPos+1))
				if newPos > 0 && newPos <= len(instructions) {
					outputFileTemp.WriteString(fmt.Sprintf(" INST:%s", instructions[newPos].String()))
					closestCase := -1
					j := 0
					for j = i; j < len(instructions); j++ {
						op := instructions[j].Op
						if op == arm64asm.B {
							if a, c := instructions[j].Args[0].(arm64asm.PCRel); c {
								if a < 0 {
									break
								}
								j += int(a) / 4
							}
						}
						casePos := -1
						for a, c := range switchInfo.Cases {
							if c.IsInside(int64(j)) {
								casePos = a
								break
							}
						}
						if casePos != -1 && casePos != switchIdx {
							closestCase = casePos
							break
						}
						//if insideCase {
						//	if lastCase == 0 {
						//		lastCase = casePos
						//	}
						//	closestCase = lastCase + 1 + switchIdx
						//	break
						//}
						//lastCase = casePos
					}
					//for ci, caseInfo := range switchInfo.Cases {
					//	off := int(caseInfo.Offset())
					//	if newPos <= off && newPos >= off-int(caseInfo.Regions[caseInfo.Offset()]) {
					//		closestCase = ci + 1
					//		break
					//	}
					//}
					if closestCase != -1 {
						outputFileTemp.WriteString(fmt.Sprintf(", CLOSEST CASE %d, DISTANCE=%d", closestCase+1, j-i))
					}
				}
			}
		}
		outputFileTemp.WriteString("\n")
		if wasInside && !isInside {
			outputFileTemp.WriteString("// END CASE\n\n\n")
		}
		wasInside = isInside
		clearASMFIle.WriteString(instruction.String() + "\n")
	}
	fmt.Println(foundConstructor)
	_ = os.WriteFile("debug.asm", outputFileTemp.Bytes(), 0644)
	_ = os.WriteFile("clear.asm", clearASMFIle.Bytes(), 0644)
}

package ios

import (
	"TLExtractor/ios/types"
	"github.com/Laky-64/goswift/proxy"
	"strings"
)

func extractObject(file *proxy.Context, obj types.ObjectInfo) error {
	//var tl schemeTypes.TLBase
	var nameBuilder strings.Builder
	if len(obj.Package) > 0 {
		nameBuilder.WriteString(obj.Package)
		nameBuilder.WriteString(".")
	}
	nameBuilder.WriteString(obj.Name)
	//if obj.Name == "birthday" {
	//	for i, instruction := range instructions {
	//		if instruction.Op == arm64asm.MOVK {
	//			shiftInfo := reflect.ValueOf(instruction.Args[1])
	//			shift := shiftInfo.FieldByName("shift").Uint()
	//			lowValue := uint64(instructions[i-1].Args[1].(arm64asm.Imm).Imm)
	//			highValue := shiftInfo.FieldByName("imm").Uint()
	//			value := highValue<<shift | lowValue
	//			fmt.Println(fmt.Sprintf("================%x=================", value))
	//		}
	//		fmt.Println(instruction)
	//	}
	//}
	return nil
}

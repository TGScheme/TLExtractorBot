package types

import "golang.org/x/arch/arm64/arm64asm"

type ObjectInfo struct {
	Name         string
	Package      string
	Address      uint64
	Instructions []arm64asm.Inst
	IsMethod     bool
	IsSecret     bool
	Parameters   string
	CaseOffset   int
	Layer        int
}

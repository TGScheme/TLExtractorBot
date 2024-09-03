package types

import "golang.org/x/arch/arm64/arm64asm"

type SwitchKind int

const (
	SwitchKindUnknown SwitchKind = iota
	SwitchCBZKind
	SwitchEQKind
	SwitchTBNZKind
	SwitchTBZKind
)

type Switch struct {
	Cases []Case
	Kind  SwitchKind
	Reg   arm64asm.Reg
}

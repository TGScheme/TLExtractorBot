package types

type ValidateType int

const (
	IsInt ValidateType = iota
	IsFloat
	IsBool
	IsURL
	NoCheck
)

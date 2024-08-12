package types

type CheckType int

const (
	InitCheck CheckType = iota
	SubmitCheck
	FinalCheck
)

package types

type CBZOrderKind int

const (
	NormalOrder CBZOrderKind = iota
	ReverseOrder
)

type CBZInfo struct {
	Order       CBZOrderKind
	Offset      int
	FirstOffset int64
}

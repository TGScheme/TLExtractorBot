package types

type TLInterface interface {
	Constructor() string
	SetConstructor(string)
	Package() string
	Parameters() []Parameter
	SetParameters([]Parameter)
	Result() string
	SetResult(string)
	IsMethod() bool
	GetLayer() int
	SetLayer(int)
}

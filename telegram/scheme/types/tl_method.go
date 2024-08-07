package types

type TLMethod struct {
	TLBase
	Method string `json:"method"`
}

func (tl *TLMethod) Package() string {
	return tl.Method
}

func (tl *TLMethod) IsMethod() bool {
	return true
}

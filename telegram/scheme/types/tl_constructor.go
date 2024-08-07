package types

type TLConstructor struct {
	TLBase
	Predicate string `json:"predicate"`
}

func (tl *TLConstructor) Package() string {
	return tl.Predicate
}

func (tl *TLConstructor) IsMethod() bool {
	return false
}

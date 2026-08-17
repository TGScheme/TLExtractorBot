package types

type TLSchemeInterface interface {
	GetConstructors() []TLInterface
	GetMethods() []TLInterface
}

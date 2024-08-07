package types

type TLScheme struct {
	Constructors []*TLConstructor `json:"constructors"`
	Methods      []*TLMethod      `json:"methods"`
}

func (scheme TLScheme) GetConstructors() []TLInterface {
	var objs []TLInterface
	for _, obj := range scheme.Constructors {
		objs = append(objs, obj)
	}
	return objs
}

func (scheme TLScheme) GetMethods() []TLInterface {
	var objs []TLInterface
	for _, obj := range scheme.Methods {
		objs = append(objs, obj)
	}
	return objs
}

package types

type RawClass struct {
	Name        string
	Prefix      string
	Content     []LineInfo
	Vars        map[string]string
	Package     string
	ParentClass string
	ParentLink  *RawClass
}

func (class *RawClass) FullName() string {
	if len(class.Package) > 0 {
		return class.Package + "." + class.Name
	}
	return class.Name
}

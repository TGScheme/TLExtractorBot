package types

type AstClass struct {
	Vars           map[string]*AstVar
	ExtendsName    string
	ExtendsPackage string
	Functions      map[string]string
}

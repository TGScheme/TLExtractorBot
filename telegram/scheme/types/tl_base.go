package types

type TLBase struct {
	ID     string      `json:"id"`
	Params []Parameter `json:"params"`
	Type   string      `json:"type"`
	Layer  int         `json:"layer"`
}

func (tl *TLBase) Constructor() string {
	return tl.ID
}

func (tl *TLBase) SetConstructor(constructor string) {
	tl.ID = constructor
}

func (tl *TLBase) Parameters() []Parameter {
	return tl.Params
}

func (tl *TLBase) SetParameters(params []Parameter) {
	tl.Params = params
}

func (tl *TLBase) Result() string {
	return tl.Type
}

func (tl *TLBase) SetResult(result string) {
	tl.Type = result
}

func (tl *TLBase) GetLayer() int {
	return tl.Layer
}

func (tl *TLBase) SetLayer(layer int) {
	tl.Layer = layer
}

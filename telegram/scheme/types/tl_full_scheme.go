package types

type TLFullScheme struct {
	MainApi TLScheme `json:"main_api"`
	E2EApi  TLScheme `json:"e2e_api"`
	Layer   int      `json:"layer"`
	IsSync  bool     `json:"-"`
}

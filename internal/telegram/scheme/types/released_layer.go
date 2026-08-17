package types

type ReleasedLayer struct {
	Constructors []ReleasedConstructor `json:"constructors"`
	Methods      []ReleasedConstructor `json:"methods"`
}

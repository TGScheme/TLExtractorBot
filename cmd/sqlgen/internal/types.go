package internal

type SQLCConfig struct {
	Catalog Catalog `json:"catalog"`
	Queries []Query `json:"queries"`
}

type Catalog struct {
	Schemas []Schema `json:"schemas"`
}

type Query struct {
	Text     string   `json:"text"`
	Name     string   `json:"name"`
	Cmd      string   `json:"cmd"`
	Columns  []Column `json:"columns"`
	Params   []Param  `json:"params"`
	Comments []string `json:"comments"`
	FileName string   `json:"filename"`
}

type Param struct {
	Number int    `json:"number"`
	Column Column `json:"column"`
}

type Schema struct {
	Name   string  `json:"name"`
	Tables []Table `json:"tables"`
	Enums  []Enum  `json:"enums"`
}

type Table struct {
	Rel struct {
		Name string `json:"name"`
	} `json:"rel"`

	Columns []Column `json:"columns"`
}

type Enum struct {
	Name string   `json:"name"`
	Vals []string `json:"vals"`
}

type Column struct {
	Name    string `json:"name"`
	NotNull bool   `json:"not_null"`
	IsArray bool   `json:"is_array"`
	Table   struct {
		Name string `json:"name"`
	} `json:"table"`
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

type QueryOptions struct {
	Cache     CacheOptions
	Order     []string
	Exclude   []string
	Returning string
}

type CacheOptions struct {
	Allow     bool
	Kind      string
	Key       string
	KeyColumn *Column
	Table     string
	VersionBy *VersionByOptions
	Fields    []string
	TTL       int64
}

type VersionByOptions struct {
	Column *Column
	Table  string
}

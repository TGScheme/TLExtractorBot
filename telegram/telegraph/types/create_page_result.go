package types

type CreatePageResult struct {
	OK     bool `json:"ok"`
	Result struct {
		URL string `json:"url"`
	}
}

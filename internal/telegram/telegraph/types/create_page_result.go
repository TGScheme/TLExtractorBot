package types

type CreatePageResult struct {
	OK     bool `json:"ok"`
	Result struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}
}

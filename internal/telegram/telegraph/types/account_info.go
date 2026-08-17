package types

type AccountInfo struct {
	OK     bool `json:"ok"`
	Result struct {
		AuthorName string `json:"author_name"`
		ShortName  string `json:"short_name"`
		AuthorURL  string `json:"author_url"`
	} `json:"result"`
}

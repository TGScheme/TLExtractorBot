package types

type CreateResult struct {
	OK     bool `json:"ok"`
	Result struct {
		AccessToken string `json:"access_token"`
	} `json:"result"`
}

package types

type Completions struct {
	Messages    []Messages `json:"messages"`
	Model       string     `json:"model"`
	Temperature int        `json:"temperature"`
	TopP        int        `json:"top_p"`
	Stream      bool       `json:"stream"`
}

package types

type CreatePageRequest struct {
	AuthorName  string `json:"author_name"`
	AuthorURL   string `json:"author_url"`
	AccessToken string `json:"access_token"`
	Title       string `json:"title"`
	Content     []Node `json:"content"`
}

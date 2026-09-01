package bot

import "github.com/TGScheme/TLExtractorBot/internal/consts"

func (ctx *Client) NextChannelPost(username string, after int) (*ChannelPost, error) {
	posts, err := ctx.fetchChannelPosts(username, after+1, consts.ChannelPostWindow)
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, nil
	}
	return posts[0], nil
}

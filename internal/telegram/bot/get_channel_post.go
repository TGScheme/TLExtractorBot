package bot

func (ctx *Client) GetChannelPost(username string, id int) (*ChannelPost, error) {
	posts, err := ctx.fetchChannelPosts(username, id, 1)
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, nil
	}
	return posts[0], nil
}

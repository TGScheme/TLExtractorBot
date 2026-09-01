package bot

import "github.com/TGScheme/TLExtractorBot/internal/consts"

func (ctx *Client) LatestChannelPost(username string) (int, error) {
	low, high := 0, 1
	for {
		found, err := ctx.anyChannelPostIn(username, high)
		if err != nil {
			return 0, err
		}
		if !found {
			break
		}
		low = high
		high *= 2
	}
	for low+1 < high {
		middle := (low + high) / 2
		found, err := ctx.anyChannelPostIn(username, middle)
		if err != nil {
			return 0, err
		}
		if found {
			low = middle
		} else {
			high = middle
		}
	}
	posts, err := ctx.fetchChannelPosts(username, low, consts.ChannelPostWindow)
	if err != nil || len(posts) == 0 {
		return low, err
	}
	return posts[len(posts)-1].ID, nil
}

func (ctx *Client) anyChannelPostIn(username string, from int) (bool, error) {
	posts, err := ctx.fetchChannelPosts(username, from, consts.ChannelPostWindow)
	return len(posts) > 0, err
}

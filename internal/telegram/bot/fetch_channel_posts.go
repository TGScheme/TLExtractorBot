package bot

import "github.com/gotd/td/tg"

func (ctx *Client) fetchChannelPosts(username string, from, count int) ([]*ChannelPost, error) {
	api, channel, err := ctx.resolveChannel(username)
	if err != nil {
		return nil, err
	}
	ids := make([]tg.InputMessageClass, 0, count)
	for id := from; id < from+count; id++ {
		if id < 1 {
			continue
		}
		ids = append(ids, &tg.InputMessageID{ID: id})
	}
	if len(ids) == 0 {
		return nil, nil
	}
	result, err := api.ChannelsGetMessages(ctx.mtProtoCtx, &tg.ChannelsGetMessagesRequest{
		Channel: channel,
		ID:      ids,
	})
	if err != nil {
		return nil, err
	}
	channelMessages, ok := result.(*tg.MessagesChannelMessages)
	if !ok {
		return nil, nil
	}
	var posts []*ChannelPost
	for _, entry := range channelMessages.Messages {
		message, isMessage := entry.(*tg.Message)
		if !isMessage {
			continue
		}
		post := &ChannelPost{ID: message.ID, Text: message.Message}
		if media, hasMedia := message.Media.(*tg.MessageMediaDocument); hasMedia {
			if document, isDocument := media.Document.(*tg.Document); isDocument {
				post.Document = document
			}
		}
		posts = append(posts, post)
	}
	return posts, nil
}

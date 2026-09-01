package bot

import (
	"fmt"

	"github.com/gotd/td/tg"
)

func (ctx *Client) resolveChannel(username string) (*tg.Client, *tg.InputChannel, error) {
	api, err := ctx.mtProtoClient()
	if err != nil {
		return nil, nil, err
	}
	ctx.mtProtoMutex.RLock()
	cached, ok := ctx.channels[username]
	ctx.mtProtoMutex.RUnlock()
	if ok {
		return api, cached, nil
	}
	resolved, err := api.ContactsResolveUsername(ctx.mtProtoCtx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, nil, err
	}
	if len(resolved.Chats) == 0 {
		return nil, nil, fmt.Errorf("@%s resolved to no chat", username)
	}
	channel, ok := resolved.Chats[0].(*tg.Channel)
	if !ok {
		return nil, nil, fmt.Errorf("@%s is a %T, not a channel", username, resolved.Chats[0])
	}
	input := &tg.InputChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}
	ctx.mtProtoMutex.Lock()
	ctx.channels[username] = input
	ctx.mtProtoMutex.Unlock()
	return api, input, nil
}

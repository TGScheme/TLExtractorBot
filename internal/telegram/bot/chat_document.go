package bot

import (
	"fmt"

	"github.com/gotd/td/tg"
)

func (ctx *Client) ChatDocument(messageID int) (*tg.Document, error) {
	api, err := ctx.mtProtoClient()
	if err != nil {
		return nil, err
	}
	result, err := api.MessagesGetMessages(ctx.mtProtoCtx, []tg.InputMessageClass{&tg.InputMessageID{ID: messageID}})
	if err != nil {
		return nil, err
	}
	var messages []tg.MessageClass
	switch box := result.(type) {
	case *tg.MessagesMessages:
		messages = box.Messages
	case *tg.MessagesMessagesSlice:
		messages = box.Messages
	}
	for _, entry := range messages {
		message, isMessage := entry.(*tg.Message)
		if !isMessage || message.ID != messageID {
			continue
		}
		if media, hasMedia := message.Media.(*tg.MessageMediaDocument); hasMedia {
			if document, isDocument := media.Document.(*tg.Document); isDocument {
				return document, nil
			}
		}
	}
	return nil, fmt.Errorf("message %d carries no document", messageID)
}

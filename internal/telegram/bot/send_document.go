package bot

import (
	"fmt"

	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
)

const maxUploadSize = 50 << 20

func (ctx *Client) SendDocument(chatID int64, name string, data []byte, caption string) error {
	if len(data) > maxUploadSize {
		return fmt.Errorf("%s is %d bytes, over the %d byte upload limit", name, len(data), maxUploadSize)
	}
	_, err := ctx.client.Invoke(
		&methods.SendDocument{
			ChatID:  chatID,
			Caption: caption,
			Document: types.InputBytes{
				Name: name,
				Data: data,
			},
			ParseMode: "HTML",
		},
	)
	return err
}

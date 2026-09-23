package bot

import (
	"github.com/GoBotApiOfficial/gobotapi/methods"
	"github.com/GoBotApiOfficial/gobotapi/types"
	rawTypes "github.com/GoBotApiOfficial/gobotapi/types/raw"
)

const bannerMediaID = "banner"

const BannerSource = "tg://photo?id=" + bannerMediaID

type richUpload struct {
	*methods.SendRichMessage
	files map[string]rawTypes.InputFile
}

func (method *richUpload) Files() map[string]rawTypes.InputFile {
	return method.files
}

func (ctx *Client) sendRich(method *methods.SendRichMessage, banner []byte) error {
	if banner == nil {
		_, err := ctx.client.Invoke(method)
		return err
	}
	method.RichMessage.Media = []types.InputRichMessageMedia{{
		ID:    bannerMediaID,
		Media: types.InputMediaPhoto{Media: types.InputURL("attach://" + bannerMediaID)},
	}}
	_, err := ctx.client.Invoke(&richUpload{
		SendRichMessage: method,
		files: map[string]rawTypes.InputFile{
			bannerMediaID: types.InputBytes{Name: bannerMediaID + ".png", Data: banner},
		},
	})
	return err
}

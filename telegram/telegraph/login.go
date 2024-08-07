package telegraph

import (
	"TLExtractor/consts"
	"TLExtractor/http"
	"TLExtractor/io"
	"TLExtractor/telegram/telegraph/types"
	"TLExtractor/utils"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

func Login() (*Context, error) {
	var ctx Context
	haveToken := false
	if len(utils.LocalStorage.BannerURL) == 0 {
		mediaUrl, err := upload(consts.Resources["banner.png"], "image/png")
		if err != nil {
			return nil, err
		}
		utils.LocalStorage.BannerURL = mediaUrl
		if err = utils.LocalStorage.Commit(); err != nil {
			return nil, err
		}
	}
	if len(utils.CredentialsStorage.TelegraphToken) == 0 {
		fmt.Print("Do you have a telegraph token? (y/n): ")
		var answer string
		_ = io.Scanln(&answer)
		haveToken = slices.Contains([]string{"y", "yes"}, strings.ToLower(answer))
	}
	for {
		if len(utils.CredentialsStorage.TelegraphToken) == 0 {
			if haveToken {
				fmt.Print("Enter telegraph token: ")
				_ = io.Scanln(&utils.CredentialsStorage.TelegraphToken)
			} else {
				var authorName, shortName, authorUrl string
				fmt.Print("Author name: ")
				_ = io.Scanln(&authorName)
				fmt.Print("Short name: ")
				_ = io.Scanln(&shortName)
				fmt.Print("Author url: ")
				_ = io.Scanln(&authorUrl)
				res := http.ExecuteRequest(
					fmt.Sprintf(
						"%s/createAccount?short_name=%s&author_name=%s&author_url=%s",
						consts.TelegraphApi,
						url.PathEscape(shortName),
						url.PathEscape(authorName),
						url.PathEscape(authorUrl),
					),
				)
				if res.Error != nil {
					return nil, res.Error
				}
				var createRes types.CreateResult
				err := json.Unmarshal(res.Read(), &createRes)
				if err != nil {
					return nil, err
				}
				utils.CredentialsStorage.TelegraphToken = createRes.Result.AccessToken
				if err = utils.CredentialsStorage.Commit(); err != nil {
					return nil, err
				}
				fmt.Println()
				fmt.Println("Your token is:", utils.CredentialsStorage.TelegraphToken)
				fmt.Println("Please save it somewhere safe, you will not be able to see it again.")
				fmt.Println()
			}
		}
		res := http.ExecuteRequest(
			fmt.Sprintf(
				"%s/getAccountInfo?access_token=%s",
				consts.TelegraphApi,
				url.PathEscape(utils.CredentialsStorage.TelegraphToken),
			),
		)
		if res.Error != nil {
			return nil, res.Error
		}
		var authRes types.AccountInfo
		err := json.Unmarshal(res.Read(), &authRes)
		if err != nil {
			return nil, err
		}
		if authRes.OK {
			ctx.accountInfo = authRes
			break
		} else {
			utils.CredentialsStorage.TelegraphToken = ""
			utils.CrashLog(consts.InvalidToken, false)
		}
	}
	return &ctx, nil
}

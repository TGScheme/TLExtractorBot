package telegraph

import (
	"TLExtractor/assets"
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/http"
	"TLExtractor/io"
	"TLExtractor/logging"
	"TLExtractor/telegram/telegraph/types"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

func init() {
	haveToken := false
	Client = &context{}
	if len(environment.LocalStorage.BannerURL) == 0 {
		mediaUrl, err := upload(assets.Resources["banner.png"], "image/png")
		if err != nil {
			logging.Fatal(err)
		}
		environment.LocalStorage.BannerURL = mediaUrl
		environment.LocalStorage.Commit()
	}
	if len(environment.CredentialsStorage.TelegraphToken) == 0 {
		fmt.Print("Do you have a telegraph token? (y/n): ")
		var answer string
		_ = io.Scanln(&answer)
		haveToken = slices.Contains([]string{"y", "yes"}, strings.ToLower(answer))
	}
	for {
		if len(environment.CredentialsStorage.TelegraphToken) == 0 {
			if haveToken {
				fmt.Print("Enter telegraph token: ")
				_ = io.Scanln(&environment.CredentialsStorage.TelegraphToken)
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
					logging.Fatal(res.Error)
				}
				var createRes types.CreateResult
				err := json.Unmarshal(res.Read(), &createRes)
				if err != nil {
					logging.Fatal(err)
				}
				environment.CredentialsStorage.TelegraphToken = createRes.Result.AccessToken
				environment.CredentialsStorage.Commit()
				logging.Info(fmt.Sprintf("Your token is: %s", environment.CredentialsStorage.TelegraphToken))
				logging.Warn("Please save it somewhere safe, you will not be able to see it again.")
			}
		}
		res := http.ExecuteRequest(
			fmt.Sprintf(
				"%s/getAccountInfo?access_token=%s",
				consts.TelegraphApi,
				url.PathEscape(environment.CredentialsStorage.TelegraphToken),
			),
		)
		if res.Error != nil {
			logging.Fatal(res.Error)
		}
		var authRes types.AccountInfo
		err := json.Unmarshal(res.Read(), &authRes)
		if err != nil {
			logging.Fatal(err)
		}
		if authRes.OK {
			Client.accountInfo = authRes
			break
		} else {
			environment.CredentialsStorage.TelegraphToken = ""
			logging.Error(consts.InvalidToken)
		}
	}
}

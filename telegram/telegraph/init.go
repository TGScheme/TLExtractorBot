package telegraph

import (
	"TLExtractor/assets"
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/telegram/telegraph/types"
	"TLExtractor/tui"
	tuiTypes "TLExtractor/tui/types"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/Laky-64/http"
	"github.com/charmbracelet/huh"
	"net/url"
)

func init() {
	Client = &context{}
	if len(environment.LocalStorage.BannerURL) == 0 {
		mediaUrl, err := upload(assets.Resources["banner.png"], "image/png")
		if err != nil {
			gologging.Fatal(err)
		}
		environment.LocalStorage.BannerURL = mediaUrl
		environment.LocalStorage.Commit()
	}
	var alreadyHave bool
	telegraphApp := tui.NewMiniApp("telegraph")
	telegraphApp.SetFields(
		huh.NewConfirm().
			Title("Do you have a Telegraph token?").
			Description("If you have a Telegraph token, you can enter it. Otherwise, we will create a new account for you.").
			Value(&alreadyHave),
	)
	telegraphApp.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		if len(environment.CredentialsStorage.TelegraphToken) == 0 {
			return errors.New("telegraph token is required")
		}
		res, err := http.ExecuteRequest(
			fmt.Sprintf(
				"%s/getAccountInfo?access_token=%s",
				consts.TelegraphApi,
				url.PathEscape(environment.CredentialsStorage.TelegraphToken),
			),
		)
		if err != nil {
			return err
		}
		var authRes types.AccountInfo
		err = json.Unmarshal(res.Body, &authRes)
		if err != nil {
			return err
		}
		if !authRes.OK {
			environment.CredentialsStorage.TelegraphToken = ""
			return consts.InvalidToken
		}
		Client.accountInfo = authRes
		return nil
	}, tuiTypes.InitCheck, tuiTypes.FinalCheck)

	var authorName, shortName, authorUrl string
	registerPage := telegraphApp.NewAppPage()
	registerPage.SetLoadingMessage("Creating a new Telegraph account...")
	registerPage.SetFields(
		huh.NewInput().
			Title("Author name").
			Description("Enter the name of the author of the Telegraph account.").
			Value(&authorName).
			Validate(tui.Validate("Author name", tuiTypes.NoCheck)),
		huh.NewInput().
			Title("Short name").
			Description("Enter a short name for your Telegraph account.").
			Value(&shortName).
			Validate(tui.Validate("Short name", tuiTypes.NoCheck)),
		huh.NewInput().
			Title("Author URL").
			Description("Enter the URL of the author of the Telegraph account.").
			Value(&authorUrl).
			Validate(tui.Validate("Author URL", tuiTypes.IsURL)),
	)
	registerPage.HideFunc(func() bool {
		return alreadyHave
	})
	registerPage.SetCheckFunc(func(checkType tuiTypes.CheckType) error {
		res, err := http.ExecuteRequest(
			fmt.Sprintf(
				"%s/createAccount?short_name=%s&author_name=%s&author_url=%s",
				consts.TelegraphApi,
				url.PathEscape(shortName),
				url.PathEscape(authorName),
				url.PathEscape(authorUrl),
			),
		)
		if err != nil {
			return err
		}
		var createRes types.CreateResult
		err = json.Unmarshal(res.Body, &createRes)
		if err != nil {
			return err
		}
		environment.CredentialsStorage.TelegraphToken = createRes.Result.AccessToken
		environment.CredentialsStorage.Commit()
		return nil
	}, tuiTypes.SubmitCheck)

	confirmTokenSaving := telegraphApp.NewAppPage()
	confirmTokenSaving.SetFields(
		huh.NewConfirm().
			Title("Save your Telegraph token").
			DescriptionFunc(func() string {
				return fmt.Sprintf(
					"Your Telegraph token is: %s\nPlease save it somewhere safe, you will not be able to see it again.",
					environment.CredentialsStorage.TelegraphToken,
				)
			}, nil).
			Affirmative("Continue").
			Negative(""),
	)
	confirmTokenSaving.HideFunc(func() bool {
		return alreadyHave
	})

	loginApp := telegraphApp.NewAppPage()
	loginApp.SetLoadingMessage("Logging in to Telegraph...")
	loginApp.SetFields(
		huh.NewInput().
			Title("Telegraph token").
			Description("Enter the Telegraph token.").
			Value(&environment.CredentialsStorage.TelegraphToken).
			EchoMode(huh.EchoModePassword).
			Validate(tui.Validate("Telegraph token", tuiTypes.NoCheck)),
	)
	loginApp.HideFunc(func() bool {
		return !alreadyHave
	})
}

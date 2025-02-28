package github

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/tui"
	"TLExtractor/tui/types"
	"context"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/charmbracelet/huh"
	"github.com/google/go-github/v69/github"
	"net/http"
	"os"
	"path"
	"strconv"
)

func init() {
	Client = &clientContext{
		ctx: context.Background(),
	}
	var filePath, appId, installationId string
	githubApp := tui.NewMiniApp("github")
	githubApp.SetLoadingMessage("Logging in to GitHub...")
	githubApp.SetFields(
		huh.NewFilePicker().
			Title("Enter the path to your GitHub App PEM file:").
			Description("You can create a GitHub App and download the PEM file from the GitHub Developer Settings.").
			ShowPermissions(false).
			AllowedTypes([]string{".pem"}).
			ShowHidden(true).
			Description("You can find your GitHub App ID in the GitHub Developer Settings.").
			Value(&filePath),
		huh.NewInput().
			Title("Enter your GitHub App ID:").
			Description("You can find your GitHub App ID in the GitHub Developer Settings.").
			Placeholder("123456").
			Validate(tui.Validate("GitHub App ID", types.IsInt)).
			Value(&appId),
		huh.NewInput().
			Title("Enter your GitHub Installation ID:").
			Description("You can find your GitHub Installation ID in the GitHub Developer Settings.").
			Placeholder("12345678").
			Validate(tui.Validate("GitHub Installation ID", types.IsInt)).
			Value(&installationId),
	)
	githubApp.SetCheckFunc(func(checkType types.CheckType) error {
		githubPemPath := path.Join(environment.EnvFolder, consts.GithubPem)
		var checkPath string
		if checkType == types.InitCheck {
			checkPath = githubPemPath
		} else {
			environment.CredentialsStorage.ApplicationID, _ = strconv.Atoi(appId)
			environment.CredentialsStorage.InstallationID, _ = strconv.Atoi(installationId)
			checkPath = filePath
		}
		file, err := os.ReadFile(checkPath)
		if err != nil {
			return err
		}
		transport, err := ghinstallation.New(
			http.DefaultTransport,
			int64(environment.CredentialsStorage.ApplicationID),
			int64(environment.CredentialsStorage.InstallationID),
			file,
		)
		if err != nil {
			_ = os.Remove(githubPemPath)
			return err
		}
		if checkType != types.InitCheck {
			if err = os.WriteFile(githubPemPath, file, os.ModePerm); err != nil {
				return err
			}
		}
		Client.client = github.NewClient(&http.Client{Transport: transport})
		if _, _, err = Client.client.Users.Get(Client.ctx, "octocat"); err != nil {
			_ = os.Remove(githubPemPath)
			environment.CredentialsStorage.ApplicationID = 0
			environment.CredentialsStorage.InstallationID = 0
			return err
		}
		environment.CredentialsStorage.Commit()
		return nil
	}, types.InitCheck, types.SubmitCheck)
}

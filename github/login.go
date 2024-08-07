package github

import (
	"TLExtractor/consts"
	"TLExtractor/io"
	"TLExtractor/utils"
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v62/github"
	"net/http"
	"os"
	"path"
	"strings"
)

func Login() (*Context, error) {
	ctx := &Context{
		ctx: context.Background(),
	}
	for {
		var githubPemPath string
		var file []byte
		if _, err := os.Stat(path.Join(consts.BasePath, consts.GithubPem)); err != nil {
			fmt.Print("Enter the path to your GitHub App PEM file: ")
			_ = io.Scanln(&githubPemPath)
			file, err = os.ReadFile(strings.TrimSpace(githubPemPath))
			if err != nil {
				utils.CrashLog(fmt.Errorf("could not read file: %s", err), false)
				continue
			}
		} else {
			file, err = os.ReadFile(path.Join(consts.BasePath, consts.GithubPem))
			if err != nil {
				return nil, err
			}
		}
		if utils.CredentialsStorage.ApplicationID == 0 {
			fmt.Print("Enter your Github App ID: ")
			_ = io.Scanln(&utils.CredentialsStorage.ApplicationID)
		}
		if utils.CredentialsStorage.InstallationID == 0 {
			fmt.Print("Enter your Github Installation ID: ")
			_ = io.Scanln(&utils.CredentialsStorage.InstallationID)
		}

		transport, err := ghinstallation.New(
			http.DefaultTransport,
			utils.CredentialsStorage.ApplicationID,
			utils.CredentialsStorage.InstallationID,
			file,
		)
		if err != nil {
			utils.CrashLog(err, false)
			_ = os.Remove(path.Join(consts.BasePath, consts.GithubPem))
			continue
		}
		if err = os.WriteFile(path.Join(consts.BasePath, consts.GithubPem), file, 0644); err != nil {
			return nil, err
		}
		ctx.client = github.NewClient(&http.Client{Transport: transport})
		if _, _, err = ctx.client.Users.Get(ctx.ctx, "octocat"); err != nil {
			utils.CrashLog(err, false)
			_ = os.Remove(path.Join(consts.BasePath, consts.GithubPem))
			utils.CredentialsStorage.ApplicationID = 0
			utils.CredentialsStorage.InstallationID = 0
			continue
		}
		if err = utils.CredentialsStorage.Commit(); err != nil {
			return nil, err
		}
		break
	}
	return ctx, nil
}

package github

import (
	"TLExtractor/consts"
	"TLExtractor/environment"
	"TLExtractor/io"
	"TLExtractor/logging"
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v62/github"
	"github.com/kardianos/service"
	"net/http"
	"os"
	"path"
	"strings"
)

func init() {
	Client = &clientContext{
		ctx: context.Background(),
	}
	for {
		var githubPemPath string
		var file []byte
		if _, err := os.Stat(path.Join(consts.EnvFolder, consts.GithubPem)); err != nil {
			if !service.Interactive() {
				logging.Fatal("GitHub App PEM file is required")
			}
			fmt.Print("Enter the path to your GitHub App PEM file: ")
			_ = io.Scanln(&githubPemPath)
			file, err = os.ReadFile(strings.TrimSpace(githubPemPath))
			if err != nil {
				logging.Error(fmt.Errorf("could not read file: %s", err))
				continue
			}
		} else {
			file, err = os.ReadFile(path.Join(consts.EnvFolder, consts.GithubPem))
			if err != nil {
				logging.Fatal(err)
			}
		}
		if environment.CredentialsStorage.ApplicationID == 0 {
			if !service.Interactive() {
				logging.Fatal("GitHub App ID is required")
			}
			fmt.Print("Enter your Github App ID: ")
			_ = io.Scanln(&environment.CredentialsStorage.ApplicationID)
		}
		if environment.CredentialsStorage.InstallationID == 0 {
			if !service.Interactive() {
				logging.Fatal("GitHub Installation ID is required")
			}
			fmt.Print("Enter your Github Installation ID: ")
			_ = io.Scanln(&environment.CredentialsStorage.InstallationID)
		}

		transport, err := ghinstallation.New(
			http.DefaultTransport,
			environment.CredentialsStorage.ApplicationID,
			environment.CredentialsStorage.InstallationID,
			file,
		)
		if err != nil {
			logging.Error(err)
			_ = os.Remove(path.Join(consts.EnvFolder, consts.GithubPem))
			continue
		}
		if err = os.WriteFile(path.Join(consts.EnvFolder, consts.GithubPem), file, 0644); err != nil {
			logging.Fatal(err)
		}
		Client.client = github.NewClient(&http.Client{Transport: transport})
		if _, _, err = Client.client.Users.Get(Client.ctx, "octocat"); err != nil {
			logging.Error(err)
			_ = os.Remove(path.Join(consts.EnvFolder, consts.GithubPem))
			environment.CredentialsStorage.ApplicationID = 0
			environment.CredentialsStorage.InstallationID = 0
			continue
		}
		environment.CredentialsStorage.Commit()
		break
	}
}

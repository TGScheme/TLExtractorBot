package github

import (
	"context"
	"net/http"
	"os"

	"github.com/TGScheme/TLExtractorBot/internal/config"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v72/github"
)

type Client struct {
	ctx       context.Context
	client    *github.Client
	repoOwner string
	repoName  string
}

func New(cfg *config.Config) (*Client, error) {
	key, err := os.ReadFile(cfg.GitHubPrivateKeyPath)
	if err != nil {
		return nil, err
	}
	transport, err := ghinstallation.New(
		http.DefaultTransport,
		cfg.GitHubAppID,
		cfg.GitHubInstallationID,
		key,
	)
	if err != nil {
		return nil, err
	}
	c := &Client{
		ctx:       context.Background(),
		client:    github.NewClient(&http.Client{Transport: transport}),
		repoOwner: cfg.SchemeRepoOwner,
		repoName:  cfg.SchemeRepoName,
	}
	if _, _, err = c.client.Users.Get(c.ctx, "octocat"); err != nil {
		return nil, err
	}
	return c, nil
}

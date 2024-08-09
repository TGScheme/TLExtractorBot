package github

import "github.com/google/go-github/v62/github"

func (ctx *clientContext) GetLastRelease(repoOwner, repoName string) (*github.RepositoryRelease, error) {
	release, _, err := ctx.client.Repositories.GetLatestRelease(ctx.ctx, repoOwner, repoName)
	if err != nil {
		return nil, err
	}
	return release, nil
}

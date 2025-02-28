package github

import "github.com/google/go-github/v69/github"

func (ctx *clientContext) GetCommits(repoOwner, repoName, path string) ([]*github.RepositoryCommit, error) {
	commits, resp, err := ctx.client.Repositories.ListCommits(ctx.ctx, repoOwner, repoName, &github.CommitsListOptions{
		Path: path,
	})
	if err != nil && (resp == nil || resp.StatusCode != 404) {
		return nil, err
	}
	return commits, nil
}

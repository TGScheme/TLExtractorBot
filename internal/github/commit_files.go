package github

import (
	"github.com/google/go-github/v72/github"
)

func (ctx *Client) commitsFiles(files map[string]string, commitMessage string) (map[string]string, error) {
	_, contents, resp, err := ctx.client.Repositories.GetContents(ctx.ctx, ctx.repoOwner, ctx.repoName, ".", nil)
	if err != nil && (resp == nil || resp.StatusCode != 404) {
		return nil, err
	}
	alreadyExists := make(map[string]string)
	for _, content := range contents {
		alreadyExists[*content.Path] = *content.SHA
	}
	commitHashes := make(map[string]string)
	for path, content := range files {
		var res *github.RepositoryContentResponse
		if sha, ok := alreadyExists[path]; ok {
			res, _, err = ctx.client.Repositories.UpdateFile(
				ctx.ctx,
				ctx.repoOwner,
				ctx.repoName,
				path,
				&github.RepositoryContentFileOptions{
					Content: []byte(content),
					Message: &commitMessage,
					SHA:     &sha,
				},
			)
			if err != nil {
				return nil, err
			}
		} else {
			res, _, err = ctx.client.Repositories.CreateFile(
				ctx.ctx,
				ctx.repoOwner,
				ctx.repoName,
				path,
				&github.RepositoryContentFileOptions{
					Content: []byte(content),
					Message: &commitMessage,
				},
			)
			if err != nil {
				return nil, err
			}
		}
		commitHashes[path] = *res.SHA
	}
	return commitHashes, nil
}

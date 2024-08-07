package github

import (
	"TLExtractor/consts"
	"github.com/google/go-github/v62/github"
)

func (ctx *Context) commitsFiles(files map[string]string, commitMessage string) (map[string]string, error) {
	_, contents, resp, err := ctx.client.Repositories.GetContents(ctx.ctx, consts.SchemeRepoOwner, consts.SchemeRepoName, ".", nil)
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
				consts.SchemeRepoOwner,
				consts.SchemeRepoName,
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
				consts.SchemeRepoOwner,
				consts.SchemeRepoName,
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

package github

import (
	"fmt"

	"github.com/google/go-github/v72/github"

	"github.com/TGScheme/TLExtractorBot/internal/consts"
)

func (ctx *Client) CommitBanner(name string, image []byte, commitMessage string) (string, error) {
	path := fmt.Sprintf("%s/%s", consts.BannersFolder, name)
	options := &github.RepositoryContentFileOptions{
		Content: image,
		Message: &commitMessage,
	}
	file, _, resp, err := ctx.client.Repositories.GetContents(
		ctx.ctx, ctx.bannersOwner, ctx.bannersName, path, nil,
	)
	if err != nil && (resp == nil || resp.StatusCode != 404) {
		return "", err
	}
	var res *github.RepositoryContentResponse
	if file != nil && file.SHA != nil {
		options.SHA = file.SHA
		res, _, err = ctx.client.Repositories.UpdateFile(ctx.ctx, ctx.bannersOwner, ctx.bannersName, path, options)
	} else {
		res, _, err = ctx.client.Repositories.CreateFile(ctx.ctx, ctx.bannersOwner, ctx.bannersName, path, options)
	}
	if err != nil {
		return "", err
	}
	if res.Commit.SHA == nil {
		return "", fmt.Errorf("github returned no commit for %s", path)
	}
	return fmt.Sprintf(
		"%s/%s/%s/%s/%s",
		consts.GithubRawURL, ctx.bannersOwner, ctx.bannersName, *res.Commit.SHA, path,
	), nil
}

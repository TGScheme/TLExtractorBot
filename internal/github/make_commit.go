package github

import (
	"errors"
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/github/types"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
	"github.com/google/go-github/v72/github"
	"maps"
	"slices"
)

func (ctx *Client) MakeCommit(fullScheme *schemeTypes.TLFullScheme, diffs schemeTypes.DifferenceStats, commitMessage string) (*types.CommitInfo, error) {
	commitFiles := make(map[string]string)
	if diffs.MainApi.Total > 0 {
		commitFiles["main_api.tl"] = scheme.ToString(fullScheme.MainApi, fullScheme.Layer, true)
	}
	if diffs.E2EApi.Total > 0 {
		commitFiles["e2e.tl"] = scheme.ToString(fullScheme.E2EApi, fullScheme.Layer, true)
	}
	if len(commitFiles) == 0 {
		return nil, errors.New("no changes to commit")
	}
	files, err := ctx.commitsFiles(
		commitFiles,
		commitMessage,
	)
	if err != nil {
		return nil, err
	}
	hashes := slices.Collect(maps.Values(files))
	return ctx.commitInfo(hashes[len(hashes)-1], files, commitFiles), nil
}

func (ctx *Client) CurrentCommit(fullScheme *schemeTypes.TLFullScheme) (*types.CommitInfo, error) {
	commits, _, err := ctx.client.Repositories.ListCommits(
		ctx.ctx, ctx.repoOwner, ctx.repoName,
		&github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: 1}},
	)
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, errors.New("the scheme repository has no commits")
	}
	head := commits[0].GetSHA()
	contents := map[string]string{
		"main_api.tl": scheme.ToString(fullScheme.MainApi, fullScheme.Layer, true),
		"e2e.tl":      scheme.ToString(fullScheme.E2EApi, fullScheme.Layer, true),
	}
	files := make(map[string]string, len(contents))
	for file := range contents {
		files[file] = head
	}
	return ctx.commitInfo(head, files, contents), nil
}

func (ctx *Client) commitInfo(source string, files, contents map[string]string) *types.CommitInfo {
	commitInfo := &types.CommitInfo{
		FilesLines: make(map[string]string),
		SourceURL: fmt.Sprintf(
			"%s/%s/%s/tree/%s",
			consts.GithubURL,
			ctx.repoOwner,
			ctx.repoName,
			source,
		),
	}
	for file, content := range contents {
		for constructor, line := range getLines(content) {
			commitInfo.FilesLines[constructor] = fmt.Sprintf(
				"%s/%s/%s/blob/%s/%s#L%d",
				consts.GithubURL,
				ctx.repoOwner,
				ctx.repoName,
				files[file],
				file,
				line,
			)
		}
	}
	return commitInfo
}

package github

import (
	"errors"
	"fmt"
	"github.com/TGScheme/TLExtractorBot/internal/consts"
	"github.com/TGScheme/TLExtractorBot/internal/github/types"
	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
	schemeTypes "github.com/TGScheme/TLExtractorBot/internal/telegram/scheme/types"
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
	commitInfo := &types.CommitInfo{
		FilesLines: make(map[string]string),
	}
	hashes := slices.Collect(maps.Values(files))
	commitInfo.SourceURL = fmt.Sprintf(
		"%s/%s/%s/tree/%s",
		consts.GithubURL,
		ctx.repoOwner,
		ctx.repoName,
		hashes[len(hashes)-1],
	)

	for file, content := range commitFiles {
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
	return commitInfo, nil
}

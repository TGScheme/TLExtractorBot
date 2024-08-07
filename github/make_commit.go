package github

import (
	"TLExtractor/consts"
	"TLExtractor/telegram/scheme"
	schemeTypes "TLExtractor/telegram/scheme/types"
	"errors"
	"fmt"
)

func (ctx *Context) MakeCommit(fullScheme *schemeTypes.TLFullScheme, diffs schemeTypes.DifferenceStats, commitMessage string) (map[string]string, error) {
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
	var commitURLs = make(map[string]string)
	for file, content := range commitFiles {
		for constructor, line := range getLines(content) {
			commitURLs[constructor] = fmt.Sprintf(
				"%s/%s/%s/blob/%s/%s#L%d",
				consts.GithubURL,
				consts.SchemeRepoOwner,
				consts.SchemeRepoName,
				files[file],
				file,
				line,
			)
		}
	}
	return commitURLs, nil
}

package github

import "github.com/google/go-github/v69/github"

func (ctx *clientContext) GetLastRelease(repoOwner, repoName, versionLock string) (*github.RepositoryRelease, error) {
	if len(versionLock) > 0 {
		releases, _, err := ctx.client.Repositories.ListReleases(ctx.ctx, repoOwner, repoName, nil)
		if err != nil {
			return nil, err
		}
		for _, release := range releases {
			if *release.TagName == versionLock {
				return release, nil
			}
		}
	}
	release, _, err := ctx.client.Repositories.GetLatestRelease(ctx.ctx, repoOwner, repoName)
	if err != nil {
		return nil, err
	}
	return release, nil
}

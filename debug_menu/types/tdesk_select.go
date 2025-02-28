package types

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/google/go-github/v69/github"
)

type TDeskSelect struct {
	*huh.Select[string]
	Value      string
	CommitList []*github.RepositoryCommit
}

func (r *TDeskSelect) NameFormat(commit *github.RepositoryCommit) string {
	return fmt.Sprintf("%s (%s)", *commit.Commit.Message, (*commit.SHA)[:7])
}

func (r *TDeskSelect) GetCommitSha() string {
	for _, commit := range r.CommitList {
		if r.Value == r.NameFormat(commit) {
			return *commit.SHA
		}
	}
	return ""
}

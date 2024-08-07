package github

import (
	"context"
	"github.com/google/go-github/v62/github"
)

type Context struct {
	ctx    context.Context
	client *github.Client
}

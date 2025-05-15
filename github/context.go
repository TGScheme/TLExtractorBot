package github

import (
	"context"
	"github.com/google/go-github/v72/github"
)

var Client *clientContext

type clientContext struct {
	ctx    context.Context
	client *github.Client
}

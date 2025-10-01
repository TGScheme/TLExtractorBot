package gemini

import (
	"context"

	"google.golang.org/genai"
)

var Client *clientContext

type clientContext struct {
	ctx             context.Context
	generativeModel *genai.Model
	apiClient       *genai.Client
}

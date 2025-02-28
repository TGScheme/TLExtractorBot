package gemini

import (
	"context"
	"github.com/google/generative-ai-go/genai"
)

var Client *clientContext

type clientContext struct {
	ctx             context.Context
	generativeModel *genai.GenerativeModel
	apiClient       *genai.Client
}

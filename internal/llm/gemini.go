package llm

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"google.golang.org/genai"
)

type geminiProvider struct {
	client *genai.Client
}

func newGeminiProvider(ctx context.Context, token string) (*geminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: token,
		HTTPOptions: genai.HTTPOptions{
			RetryOptions: &genai.HTTPRetryOptions{Attempts: new(int32(1))},
		},
	})
	if err != nil {
		return nil, err
	}
	return &geminiProvider{client: client}, nil
}

func (p *geminiProvider) generate(ctx context.Context, request generation) (string, usage, error) {
	resp, err := p.client.Models.GenerateContent(
		ctx,
		request.model,
		[]*genai.Content{{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{{Text: request.prompt}},
		}},
		&genai.GenerateContentConfig{
			Temperature:      new(float32(0.35)),
			TopP:             new(float32(0.95)),
			MaxOutputTokens:  65536,
			ResponseMIMEType: "application/json",
			ResponseSchema:   changelogSchema,
			SystemInstruction: &genai.Content{
				Role:  genai.RoleUser,
				Parts: []*genai.Part{{Text: request.system}},
			},
		},
	)
	if err != nil {
		return "", usage{}, err
	}
	var used usage
	if meta := resp.UsageMetadata; meta != nil {
		used = usage{
			prompt:   int64(meta.PromptTokenCount),
			thinking: int64(meta.ThoughtsTokenCount),
			output:   int64(meta.CandidatesTokenCount),
		}
	}
	if len(resp.Candidates) == 0 {
		return "", used, errors.New("no candidates")
	}
	return resp.Text(), used, nil
}

func (p *geminiProvider) models(ctx context.Context) ([]Model, error) {
	var models []Model
	for model, err := range p.client.Models.All(ctx) {
		if err != nil {
			return nil, err
		}
		id := strings.TrimPrefix(model.Name, "models/")
		if strings.HasPrefix(id, "gemini") {
			models = append(models, Model{ID: id, Kind: KindGemini})
		}
	}
	return models, nil
}

func (p *geminiProvider) quotaPause(err error) (time.Duration, bool) {
	var apiErr genai.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusTooManyRequests {
		return 0, false
	}
	for _, detail := range apiErr.Details {
		if raw, ok := detail["retryDelay"].(string); ok {
			if delay, parseErr := time.ParseDuration(raw); parseErr == nil && delay > 0 {
				return delay, true
			}
		}
	}
	return defaultQuotaPause, true
}

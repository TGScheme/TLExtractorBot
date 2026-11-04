package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const openAIRateLimitPause = time.Minute

var (
	openAIBaseURL       = "https://api.openai.com/v1"
	openAIChatModel     = regexp.MustCompile(`^(gpt-\d|o\d)`)
	openAIExcludedModel = regexp.MustCompile(`audio|realtime|tts|transcribe|image|search|instruct|embedding|-\d{4}-\d{2}-\d{2}$`)
)

type openAIProvider struct {
	token           string
	reasoningEffort string
	client          *http.Client
}

type openAIError struct {
	Status     int
	Code       string
	Message    string
	RetryAfter time.Duration
}

func (e *openAIError) Error() string {
	return fmt.Sprintf("openai error %d (%s): %s", e.Status, e.Code, e.Message)
}

func newOpenAIProvider(token, reasoningEffort string) *openAIProvider {
	return &openAIProvider{token: token, reasoningEffort: reasoningEffort, client: &http.Client{Timeout: 15 * time.Minute}}
}

func (p *openAIProvider) generate(ctx context.Context, request generation) (string, usage, error) {
	payload := map[string]any{
		"model": request.model,
		"messages": []map[string]string{
			{"role": "system", "content": request.system},
			{"role": "user", "content": request.prompt},
		},
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "changelog",
				"strict": true,
				"schema": jsonSchema(changelogSchema),
			},
		},
	}
	if p.reasoningEffort != "" {
		payload["reasoning_effort"] = p.reasoningEffort
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", usage{}, err
	}
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
				Refusal string `json:"refusal"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens            int64 `json:"prompt_tokens"`
			CompletionTokens        int64 `json:"completion_tokens"`
			CompletionTokensDetails struct {
				ReasoningTokens int64 `json:"reasoning_tokens"`
			} `json:"completion_tokens_details"`
		} `json:"usage"`
	}
	if err = p.call(ctx, http.MethodPost, "/chat/completions", body, &resp); err != nil {
		return "", usage{}, err
	}
	used := usage{
		prompt:   resp.Usage.PromptTokens,
		thinking: resp.Usage.CompletionTokensDetails.ReasoningTokens,
		output:   resp.Usage.CompletionTokens - resp.Usage.CompletionTokensDetails.ReasoningTokens,
	}
	if len(resp.Choices) == 0 {
		return "", used, errors.New("no choices")
	}
	if refusal := resp.Choices[0].Message.Refusal; refusal != "" {
		return "", used, fmt.Errorf("refused: %s", refusal)
	}
	return resp.Choices[0].Message.Content, used, nil
}

func (p *openAIProvider) models(ctx context.Context) ([]Model, error) {
	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := p.call(ctx, http.MethodGet, "/models", nil, &resp); err != nil {
		return nil, err
	}
	var models []Model
	for _, model := range resp.Data {
		if openAIChatModel.MatchString(model.ID) && !openAIExcludedModel.MatchString(model.ID) {
			models = append(models, Model{ID: model.ID, Kind: KindOpenAI})
		}
	}
	return models, nil
}

func (p *openAIProvider) quotaPause(err error) (time.Duration, bool) {
	var apiErr *openAIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusTooManyRequests {
		return 0, false
	}
	switch {
	case apiErr.Code == "insufficient_quota":
		return defaultQuotaPause, true
	case apiErr.RetryAfter > 0:
		return apiErr.RetryAfter, true
	}
	return openAIRateLimitPause, true
}

func (p *openAIProvider) call(ctx context.Context, method, path string, body []byte, out any) error {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, openAIBaseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		apiErr := &openAIError{Status: resp.StatusCode, Message: strings.TrimSpace(string(raw))}
		var payload struct {
			Error struct {
				Message string `json:"message"`
				Code    string `json:"code"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if json.Unmarshal(raw, &payload) == nil && payload.Error.Message != "" {
			apiErr.Message = payload.Error.Message
			apiErr.Code = payload.Error.Code
			if apiErr.Code == "" {
				apiErr.Code = payload.Error.Type
			}
		}
		if seconds, parseErr := strconv.Atoi(resp.Header.Get("Retry-After")); parseErr == nil && seconds > 0 {
			apiErr.RetryAfter = time.Duration(seconds) * time.Second
		}
		return apiErr
	}
	return json.Unmarshal(raw, out)
}

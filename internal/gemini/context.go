package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"google.golang.org/genai"
)

const defaultQuotaPause = time.Hour

type Client struct {
	ctx     context.Context
	clients []*genai.Client
	model   string

	mutex   sync.Mutex
	current int
	blocked []time.Time
}

type QuotaError struct {
	RetryAt time.Time
	Err     error
}

func (e *QuotaError) Error() string {
	return fmt.Sprintf("every gemini key is out of quota until %s: %v", e.RetryAt.Format(time.RFC3339), e.Err)
}

func (e *QuotaError) Unwrap() error { return e.Err }

type Model struct {
	ID    string
	Label string
}

func New(cfg *config.Config, model string) (*Client, error) {
	ctx := context.Background()
	var clients []*genai.Client
	for key := range strings.SplitSeq(cfg.GeminiToken, ",") {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		apiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: key,
			HTTPOptions: genai.HTTPOptions{
				RetryOptions: &genai.HTTPRetryOptions{Attempts: new(int32(1))},
			},
		})
		if err != nil {
			return nil, err
		}
		clients = append(clients, apiClient)
	}
	if len(clients) == 0 {
		return nil, errors.New("no gemini api key configured")
	}
	return &Client{ctx: ctx, clients: clients, model: model, blocked: make([]time.Time, len(clients))}, nil
}

func (ctx *Client) withKeys(call func(apiClient *genai.Client) error) error {
	var lastErr error
	for range ctx.clients {
		apiClient, index, ok := ctx.availableKey()
		if !ok {
			break
		}
		err := call(apiClient)
		if err == nil || !isQuotaError(err) {
			return err
		}
		lastErr = err
		pause := quotaPause(err)
		ctx.mutex.Lock()
		ctx.blocked[index] = time.Now().Add(pause)
		ctx.mutex.Unlock()
		gologging.Warn(fmt.Sprintf("gemini: key %d of %d is out of quota for %s", index+1, len(ctx.clients), pause.Round(time.Second)))
	}
	return &QuotaError{RetryAt: ctx.firstUnblock(), Err: lastErr}
}

func (ctx *Client) availableKey() (*genai.Client, int, bool) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	now := time.Now()
	for offset := range ctx.clients {
		index := (ctx.current + offset) % len(ctx.clients)
		if now.After(ctx.blocked[index]) {
			ctx.current = index
			return ctx.clients[index], index, true
		}
	}
	return nil, 0, false
}

func (ctx *Client) firstUnblock() time.Time {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	first := ctx.blocked[0]
	for _, until := range ctx.blocked[1:] {
		if until.Before(first) {
			first = until
		}
	}
	return first
}

func isQuotaError(err error) bool {
	var apiErr genai.APIError
	return errors.As(err, &apiErr) && apiErr.Code == http.StatusTooManyRequests
}

func quotaPause(err error) time.Duration {
	var apiErr genai.APIError
	if !errors.As(err, &apiErr) {
		return defaultQuotaPause
	}
	for _, detail := range apiErr.Details {
		if raw, ok := detail["retryDelay"].(string); ok {
			if delay, parseErr := time.ParseDuration(raw); parseErr == nil && delay > 0 {
				return delay
			}
		}
	}
	return defaultQuotaPause
}

func (ctx *Client) apiClient() (*genai.Client, int) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	return ctx.clients[ctx.current], ctx.current
}

func (ctx *Client) Models() ([]Model, error) {
	var models []Model
	apiClient, _ := ctx.apiClient()
	for model, err := range apiClient.Models.All(ctx.ctx) {
		if err != nil {
			return nil, err
		}
		id := strings.ReplaceAll(model.Name, "models/", "")
		if !strings.HasPrefix(id, "gemini") {
			continue
		}
		models = append(models, Model{ID: id, Label: fmt.Sprintf("%s (%s)", model.DisplayName, id)})
	}
	sort.Slice(models, func(i, j int) bool { return models[i].ID > models[j].ID })
	return models, nil
}

func (ctx *Client) SetModel(model string) { ctx.model = model }

func (ctx *Client) Model() string { return ctx.model }

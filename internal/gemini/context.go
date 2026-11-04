package gemini

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
	"google.golang.org/genai"
)

const (
	retryAttempts     = 5
	retryInitialDelay = 5.0
	retryMaxDelay     = 30.0
)

type Client struct {
	ctx     context.Context
	clients []*genai.Client
	model   string

	mutex   sync.Mutex
	current int
}

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
				RetryOptions: &genai.HTTPRetryOptions{
					Attempts:        new(int32(retryAttempts)),
					InitialDelay:    new(retryInitialDelay),
					MaxDelay:        new(retryMaxDelay),
					HTTPStatusCodes: []int32{500, 502, 503, 504},
				},
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
	return &Client{ctx: ctx, clients: clients, model: model}, nil
}

func (ctx *Client) apiClient() (*genai.Client, int) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	return ctx.clients[ctx.current], ctx.current
}

func (ctx *Client) withKeys(call func(apiClient *genai.Client) error) error {
	var err error
	for range ctx.clients {
		apiClient, index := ctx.apiClient()
		if err = call(apiClient); err == nil || !isQuotaError(err) || len(ctx.clients) == 1 {
			return err
		}
		ctx.mutex.Lock()
		if ctx.current == index {
			ctx.current = (index + 1) % len(ctx.clients)
		}
		next := ctx.current
		ctx.mutex.Unlock()
		gologging.Warn(fmt.Sprintf("gemini: key %d of %d is out of quota, switching to key %d", index+1, len(ctx.clients), next+1))
	}
	return err
}

func isQuotaError(err error) bool {
	var apiErr genai.APIError
	return errors.As(err, &apiErr) && apiErr.Code == http.StatusTooManyRequests
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

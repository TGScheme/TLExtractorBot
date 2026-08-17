package gemini

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/TGScheme/TLExtractorBot/internal/config"
	"google.golang.org/genai"
)

type Client struct {
	ctx       context.Context
	apiClient *genai.Client
	model     string
}

type Model struct {
	ID    string
	Label string
}

func New(cfg *config.Config, model string) (*Client, error) {
	ctx := context.Background()
	apiClient, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiToken})
	if err != nil {
		return nil, err
	}
	return &Client{ctx: ctx, apiClient: apiClient, model: model}, nil
}

func (ctx *Client) Models() ([]Model, error) {
	var models []Model
	for model, err := range ctx.apiClient.Models.All(ctx.ctx) {
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

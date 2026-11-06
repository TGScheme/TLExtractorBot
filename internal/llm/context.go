package llm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	"github.com/TGScheme/TLExtractorBot/internal/config"
)

const (
	KindOpenAI = "OpenAI"
	KindGemini = "Gemini"

	defaultQuotaPause = time.Hour
)

type provider interface {
	generate(ctx context.Context, request generation) (string, usage, error)
	models(ctx context.Context) ([]Model, error)
	quotaPause(err error) (time.Duration, bool)
}

type generation struct {
	model  string
	system string
	prompt string
}

type usage struct {
	prompt   int64
	thinking int64
	output   int64
}

type key struct {
	kind     string
	number   int
	provider provider
}

type Client struct {
	ctx  context.Context
	keys []key

	mutex   sync.Mutex
	blocked []time.Time
	models  map[string]string
	active  string
}

type Model struct {
	ID   string
	Kind string
}

type QuotaError struct {
	RetryAt time.Time
	Err     error
}

func (e *QuotaError) Error() string {
	return fmt.Sprintf("every key is out of quota until %s: %v", e.RetryAt.Format(time.RFC3339), e.Err)
}

func (e *QuotaError) Unwrap() error { return e.Err }

func New(cfg *config.Config, geminiModel, openAIModel string) (*Client, error) {
	ctx := context.Background()
	client := &Client{
		ctx:    ctx,
		models: map[string]string{KindGemini: geminiModel, KindOpenAI: openAIModel},
	}
	for number, token := range splitTokens(cfg.OpenAIToken) {
		client.keys = append(client.keys, key{KindOpenAI, number + 1, newOpenAIProvider(token, cfg.OpenAIReasoningEffort)})
	}
	for number, token := range splitTokens(cfg.GeminiToken) {
		gemini, err := newGeminiProvider(ctx, token)
		if err != nil {
			return nil, err
		}
		client.keys = append(client.keys, key{KindGemini, number + 1, gemini})
	}
	if len(client.keys) == 0 {
		return nil, errors.New("no llm api key configured")
	}
	client.blocked = make([]time.Time, len(client.keys))
	for _, kind := range client.Kinds() {
		if client.Model(kind) == "" {
			gologging.Warn(fmt.Sprintf("llm: %s keys are set but no %s model is selected, pick one with /model", kind, kind))
		}
	}
	return client, nil
}

func splitTokens(raw string) []string {
	var tokens []string
	for token := range strings.SplitSeq(raw, ",") {
		if token = strings.TrimSpace(token); token != "" {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func (ctx *Client) Kinds() []string {
	var kinds []string
	for _, k := range ctx.keys {
		if len(kinds) == 0 || kinds[len(kinds)-1] != k.kind {
			kinds = append(kinds, k.kind)
		}
	}
	return kinds
}

func (ctx *Client) ActiveKind() string {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	if ctx.active != "" {
		return ctx.active
	}
	for _, k := range ctx.keys {
		if ctx.models[k.kind] != "" {
			return k.kind
		}
	}
	return ""
}

func (ctx *Client) Model(kind string) string {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	return ctx.models[kind]
}

func (ctx *Client) SetModel(kind, model string) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.models[kind] = model
}

func (ctx *Client) Models() ([]Model, error) {
	var models []Model
	for _, kind := range ctx.Kinds() {
		for _, k := range ctx.keys {
			if k.kind != kind {
				continue
			}
			list, err := k.provider.models(ctx.ctx)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", kind, err)
			}
			sort.Slice(list, func(i, j int) bool { return list[i].ID > list[j].ID })
			models = append(models, list...)
			break
		}
	}
	return models, nil
}

func (ctx *Client) withKeys(call func(k key, model string) error) (string, error) {
	var lastErr error
	onlyQuota, tried := true, false
	failed := make(map[string]bool)
	for index, k := range ctx.keys {
		model := ctx.Model(k.kind)
		if model == "" || failed[k.kind] || !ctx.available(index) {
			continue
		}
		tried = true
		ctx.mutex.Lock()
		ctx.active = k.kind
		ctx.mutex.Unlock()
		err := call(k, model)
		if err == nil {
			return model, nil
		}
		lastErr = err
		if pause, ok := k.provider.quotaPause(err); ok {
			ctx.mutex.Lock()
			ctx.blocked[index] = time.Now().Add(pause)
			ctx.mutex.Unlock()
			gologging.Warn(fmt.Sprintf("llm: %s key %d is out of quota for %s", k.kind, k.number, pause.Round(time.Second)))
			continue
		}
		onlyQuota = false
		failed[k.kind] = true
		gologging.Warn(fmt.Sprintf("llm: %s failed with %s: %v", k.kind, model, err))
	}
	if !tried && lastErr == nil {
		if retryAt, ok := ctx.firstUnblock(); ok {
			return "", &QuotaError{RetryAt: retryAt}
		}
		return "", errNoModel
	}
	if onlyQuota {
		retryAt, _ := ctx.firstUnblock()
		return "", &QuotaError{RetryAt: retryAt, Err: lastErr}
	}
	return "", lastErr
}

var errNoModel = errors.New("no model selected, pick one with /model")

func (ctx *Client) available(index int) bool {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	return time.Now().After(ctx.blocked[index])
}

func (ctx *Client) firstUnblock() (time.Time, bool) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	var first time.Time
	found := false
	for index, k := range ctx.keys {
		if ctx.models[k.kind] == "" {
			continue
		}
		if !found || ctx.blocked[index].Before(first) {
			first, found = ctx.blocked[index], true
		}
	}
	return first, found
}

// Package ai is a swappable LLM provider boundary. Finance business logic
// depends only on this interface; the mock provider keeps tests and demos
// deterministic and offline (AGENTS #76, #77).
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/savio/savio/backend/internal/platform/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Provider calls a model and returns its textual completion. Callers must
// validate/enforce JSON schemas themselves.
type Provider interface {
	Complete(ctx context.Context, system, prompt string) (string, error)
	// Mock reports whether this is the deterministic test/demo provider, in
	// which case callers should prefer deterministic rules over its canned
	// output (AGENTS #77: AI degradation must not corrupt behavior).
	Mock() bool
}

type Config interface {
	AIBaseURL() string
	AIAPIKey() string
	AIModel() string
	AITimeout() time.Duration
}

func NewProvider(cfg Config) Provider {
	base := cfg.AIBaseURL()
	key := cfg.AIAPIKey()
	if base == "" || key == "" {
		return Mock{WhenEmpty: true}
	}
	return &OpenAIProvider{cfg: cfg}
}

// Mock is a deterministic provider for tests and AI-disabled environments. It
// returns a canned JSON object keyed by a marker in the prompt so callers can
// exercise real parse/validate paths.
type Mock struct {
	WhenEmpty bool
}

func (Mock) Mock() bool { return true }

func (Mock) Complete(ctx context.Context, _, prompt string) (string, error) {
	switch {
	case strings.Contains(prompt, "BROKEN"):
		return `{"category_guess": "not valid json`, nil
	case strings.Contains(prompt, "CATEGORIZE"):
		return `{"category_guess":"Food & Dining","confidence":0.9,"matched_rule":"system"}`, nil
	case strings.Contains(prompt, "INSIGHT"):
		return `{"headline":"Spending is up this month","detail":"Expenses rose 12% vs last month.","signal":"spending_increase","related_facts":["expenses 12%","income flat"]}`, nil
	case strings.Contains(prompt, "COPILOT"):
		return `{"answer":"Your forecast projects a comfortable minimum balance this horizon.","sources":["forecast"],"caveats":["estimates only"]}`, nil
	case strings.Contains(prompt, "EXPLAIN"):
		return `{"explanation":"This scenario reduces your ending balance by the one-time expense amount.","impact":"ending_balance_reduction"}`, nil
	default:
		return `{"ok":true}`, nil
	}
}

type OpenAIProvider struct {
	cfg Config
}

func (*OpenAIProvider) Mock() bool { return false }

func (o *OpenAIProvider) Complete(ctx context.Context, system, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, o.cfg.AITimeout())
	defer cancel()

	body := map[string]any{
		"model": o.cfg.AIModel(),
		"messages": []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt},
		},
		"temperature":     0,
		"response_format": map[string]string{"type": "json_object"},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("ai: marshal: %w", err)
	}
	// Accept both an API root and a full chat/completions URL: appending
	// unconditionally would turn a user-supplied ".../chat/completions" into a
	// broken double path.
	base := strings.TrimRight(o.cfg.AIBaseURL(), "/")
	if !strings.HasSuffix(base, "/chat/completions") {
		base += "/chat/completions"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("ai: request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.cfg.AIAPIKey())
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai: provider unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return "", fmt.Errorf("ai: provider status %d", resp.StatusCode)
	}
	var parsed struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("ai: decode: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("ai: empty choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

// configAdapter adapts *config.Config to the ai.Config interface.
type configAdapter struct{ cfg *config.Config }

func (a configAdapter) AIBaseURL() string { return a.cfg.AIBaseURL }
func (a configAdapter) AIAPIKey() string  { return a.cfg.AIAPIKey }
func (a configAdapter) AIModel() string   { return a.cfg.AIModel }
func (a configAdapter) AITimeout() time.Duration {
	return a.cfg.AITimeout
}

func NewProviderFromConfig(cfg *config.Config) Provider { return NewProvider(configAdapter{cfg: cfg}) }

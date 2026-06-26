// Package ollama implements the ai.Provider interface against a local Ollama
// server. This is nock's first-class offline / air-gapped path (Milestone 4):
// inference runs in the separate Ollama process; nock only speaks HTTP to it, so
// no model is embedded and the binary stays a thin Go client.
//
// Scaffold status: Suggest returns ai.ErrNotImplemented. The real implementation
// POSTs to {BaseURL}/api/chat with a ranking prompt and parses the model's ordered
// verdict. No external SDK is required — Ollama exposes a plain HTTP/JSON API.
package ollama

import (
	"context"

	"github.com/jessn-dev/nock/internal/ai"
)

// DefaultBaseURL is Ollama's local endpoint.
const DefaultBaseURL = "http://localhost:11434"

// Config configures the Ollama provider.
type Config struct {
	BaseURL string // defaults to DefaultBaseURL
	Model   string // e.g. "llama3.1", "qwen2.5-coder"
}

// Provider talks to a local Ollama server.
type Provider struct {
	cfg Config
}

// New returns an Ollama provider, applying defaults for empty fields.
func New(cfg Config) *Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	return &Provider{cfg: cfg}
}

// Name identifies this provider.
func (p *Provider) Name() string { return "ollama" }

// Suggest is not yet implemented; the ranker falls back to fuzzy order until it is.
func (p *Provider) Suggest(ctx context.Context, intent string, candidates []ai.Candidate) ([]ai.Ranked, error) {
	return nil, ai.ErrNotImplemented
}

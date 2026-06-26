// Package openaicompat implements the ai.Provider interface against any endpoint
// that speaks the OpenAI Chat Completions API. One adapter covers a large fraction
// of backends — OpenAI itself, Groq, Together, vLLM, LocalAI, LM Studio — selected
// by BaseURL. This is why nock ships broad provider coverage with minimal code.
//
// Scaffold status: Suggest returns ai.ErrNotImplemented. The real implementation
// (Milestone 4) POSTs to {BaseURL}/chat/completions with a structured ranking
// prompt. The API key is read from the environment by the caller and never stored.
package openaicompat

import (
	"context"

	"github.com/jessn-dev/nock/internal/ai"
)

// DefaultBaseURL points at OpenAI; override for Groq, Together, vLLM, LocalAI, etc.
const DefaultBaseURL = "https://api.openai.com/v1"

// Config configures an OpenAI-compatible provider.
type Config struct {
	BaseURL string // defaults to DefaultBaseURL
	APIKey  string // may be empty for local servers (vLLM, LocalAI, LM Studio)
	Model   string // e.g. "gpt-4o-mini", "llama-3.1-70b"
}

// Provider calls an OpenAI-compatible chat endpoint.
type Provider struct {
	cfg Config
}

// New returns an OpenAI-compatible provider, applying the default base URL when unset.
func New(cfg Config) *Provider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	return &Provider{cfg: cfg}
}

// Name identifies this provider.
func (p *Provider) Name() string { return "openai-compat" }

// Suggest is not yet implemented; the ranker falls back to fuzzy order until it is.
func (p *Provider) Suggest(ctx context.Context, intent string, candidates []ai.Candidate) ([]ai.Ranked, error) {
	return nil, ai.ErrNotImplemented
}

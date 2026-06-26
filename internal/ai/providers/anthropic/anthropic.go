// Package anthropic implements the ai.Provider interface against the Anthropic
// (Claude) API — nock's default cloud ranker when high quality is wanted and the
// operator has internet plus a key.
//
// Scaffold status: Suggest returns ai.ErrNotImplemented. The real implementation
// (Milestone 4) uses github.com/anthropics/anthropic-sdk-go with a tool-use /
// structured-output ranking prompt. The API key is read from the environment
// (ANTHROPIC_API_KEY), never stored by nock.
package anthropic

import (
	"context"

	"github.com/jessn-dev/nock/internal/ai"
)

// DefaultModel favours quality; callers may set a cheaper model for bulk ranking.
const DefaultModel = "claude-opus-4-8"

// Config configures the Anthropic provider. APIKey is resolved from the environment
// by the caller; this package never persists it.
type Config struct {
	APIKey string
	Model  string // defaults to DefaultModel
}

// Provider calls the Anthropic API.
type Provider struct {
	cfg Config
}

// New returns an Anthropic provider, applying the default model when unset.
func New(cfg Config) *Provider {
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	return &Provider{cfg: cfg}
}

func (p *Provider) Name() string { return "anthropic" }

// Suggest is not yet implemented; the ranker falls back to fuzzy order until it is.
func (p *Provider) Suggest(ctx context.Context, intent string, candidates []ai.Candidate) ([]ai.Ranked, error) {
	return nil, ai.ErrNotImplemented
}

// Package ai defines nock's provider-agnostic AI ranking layer. There are two
// independent axes deliberately kept separate:
//
//  1. LLM provider (engine -> model): abstracted by the Provider interface below.
//     Swappable by config — Anthropic, OpenAI-compatible, Ollama (offline), etc.
//     The ranker depends only on this interface and never imports a vendor SDK.
//
//  2. MCP host (agent -> engine): handled in internal/mcp, provider-agnostic by
//     protocol. One MCP server works with every MCP-capable host, no per-host code.
//
// AI is always an optional upgrade. With no provider configured the engine's
// fuzzy search still ranks results, so nock works offline and key-less.
package ai

import (
	"context"
	"errors"

	"github.com/jessn-dev/nock/pkg/format"
)

// ErrNotImplemented is returned by provider stubs that are not yet wired to a real
// backend. Callers should treat it as "AI unavailable" and fall back to fuzzy order.
var ErrNotImplemented = errors.New("ai: provider not implemented yet")

// Candidate is a command offered to the ranker, carrying the engine's prior
// (fuzzy) score so a provider can blend rather than discard offline signal.
type Candidate struct {
	Command format.Command
	Prior   int
}

// Ranked is a provider's verdict on one candidate.
type Ranked struct {
	ID     string  // Command.ID
	Score  float64 // provider confidence, higher is better
	Reason string  // optional short rationale, surfaced in the UI
}

// Provider is the single seam between nock and any LLM backend. Implementations
// live in internal/ai/providers/*. Keep this interface small: ranking is the only
// capability the engine needs from a model.
type Provider interface {
	// Name identifies the provider for config, logs, and the UI (e.g. "ollama").
	Name() string

	// Suggest ranks candidates against the operator's natural-language intent.
	// Implementations must honour ctx cancellation/deadlines and must not panic on
	// an empty candidate set. Returning ErrNotImplemented signals graceful fallback.
	Suggest(ctx context.Context, intent string, candidates []Candidate) ([]Ranked, error)
}

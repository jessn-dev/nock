// Package tui is nock's terminal frontend — the default mode and the product's
// face. It is a thin shell over the engine: it renders search results, prompts for
// missing variables, and emits the resolved command to the shell. It must contain
// no search or resolution logic of its own.
//
// Scaffold status: Run returns errNotImplemented. The real implementation
// (Milestone 2) is a Bubble Tea program (charmbracelet/bubbletea + bubbles +
// lipgloss): a filterable list bound to engine.Search and a variable-prompt form
// bound to engine.Vars.
package tui

import (
	"context"
	"errors"

	"github.com/jessn-dev/nock/internal/engine"
)

var errNotImplemented = errors.New("tui: interactive UI arrives in Milestone 2 (charmbracelet/bubbletea)")

// Run starts the interactive TUI against the given engine.
func Run(ctx context.Context, e *engine.Engine) error {
	return errNotImplemented
}

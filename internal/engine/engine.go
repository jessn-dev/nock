// Package engine is the heart of nock and the single source of truth for command
// search and variable resolution. Every frontend — TUI, MCP server, future web —
// must go through this API and nothing else. No frontend may duplicate resolution
// or ranking logic; if behaviour needs to change, it changes here once.
package engine

import (
	"fmt"
	"strings"

	"github.com/jessn-dev/nock/internal/search"
	"github.com/jessn-dev/nock/internal/vars"
	"github.com/jessn-dev/nock/pkg/format"
)

// Engine holds the loaded command corpus and the variable store. Construct with New.
type Engine struct {
	commands []format.Command
	bySheet  map[string]string // command ID -> cheatsheet name, for display/provenance
	vars     *vars.Store
}

// Result is a search hit: the matched command plus its score and source sheet.
type Result struct {
	Command format.Command
	Sheet   string
	Score   int
}

// New builds an Engine from loaded cheatsheets and a variable store.
func New(sheets []format.Cheatsheet, store *vars.Store) *Engine {
	if store == nil {
		store = vars.New()
	}
	e := &Engine{
		bySheet: map[string]string{},
		vars:    store,
	}
	for _, cs := range sheets {
		for _, cmd := range cs.Commands {
			e.commands = append(e.commands, cmd)
			e.bySheet[cmd.ID] = cs.Name
		}
	}
	return e
}

// Vars exposes the underlying variable store so frontends can Set/Get bindings.
func (e *Engine) Vars() *vars.Store { return e.vars }

// Search ranks commands against a free-text query using the offline fuzzy matcher.
// The AI ranker (Milestone 4) will re-rank these Results; it does not replace them,
// so search always works with zero AI and zero network.
func (e *Engine) Search(query string) []Result {
	items := make([]search.Item, len(e.commands))
	index := make(map[string]format.Command, len(e.commands))
	for i, cmd := range e.commands {
		items[i] = search.Item{ID: cmd.ID, Text: searchText(cmd)}
		index[cmd.ID] = cmd
	}
	ranked := search.Rank(query, items)
	out := make([]Result, 0, len(ranked))
	for _, r := range ranked {
		cmd := index[r.ID]
		out = append(out, Result{Command: cmd, Sheet: e.bySheet[r.ID], Score: r.Score})
	}
	return out
}

// Resolve fills a command's <var> placeholders from the variable store, returning
// the runnable command string. Unbound variables produce an error, never a
// half-filled command.
func (e *Engine) Resolve(cmd format.Command) (string, error) {
	resolved, err := e.vars.Resolve(cmd.Command)
	if err != nil {
		return "", fmt.Errorf("engine: resolve %q: %w", cmd.ID, err)
	}
	return resolved, nil
}

// Get returns the command with the given ID, if present.
func (e *Engine) Get(id string) (format.Command, bool) {
	for _, cmd := range e.commands {
		if cmd.ID == id {
			return cmd, true
		}
	}
	return format.Command{}, false
}

// Len reports how many commands are loaded.
func (e *Engine) Len() int { return len(e.commands) }

// searchText builds the haystack the fuzzy matcher scores against: name, intent,
// description, and tags all contribute, so a query can hit any of them.
func searchText(c format.Command) string {
	parts := []string{c.Name, c.Intent, c.Description, c.Command}
	parts = append(parts, c.Tags...)
	return strings.Join(parts, " ")
}

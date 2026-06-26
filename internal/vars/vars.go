// Package vars implements nock's global variable store: set a value once
// (e.g. target=10.0.0.5) and have it fill every command template that references
// <target>. This is the "set once, fill everywhere" behaviour from arsenal.
package vars

import (
	"fmt"
	"regexp"
	"sort"
)

var placeholderRe = regexp.MustCompile(`<([a-zA-Z_][a-zA-Z0-9_]*)>`)

// Store holds variable name -> value bindings. The zero value is not usable;
// call New. A Store is not safe for concurrent writes; guard externally if shared
// across the TUI event loop and an MCP request goroutine.
type Store struct {
	values map[string]string
}

// New returns an empty Store.
func New() *Store { return &Store{values: map[string]string{}} }

// Set binds name to value, overwriting any previous binding.
func (s *Store) Set(name, value string) { s.values[name] = value }

// Get returns the value bound to name and whether it was set.
func (s *Store) Get(name string) (string, bool) {
	v, ok := s.values[name]
	return v, ok
}

// Names returns all bound variable names, sorted, for display.
func (s *Store) Names() []string {
	out := make([]string, 0, len(s.values))
	for k := range s.values {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Missing returns the variables referenced in template that are not yet bound,
// in first-seen order. Frontends use this to know what to prompt for.
func (s *Store) Missing(template string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range placeholderRe.FindAllStringSubmatch(template, -1) {
		name := m[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, ok := s.values[name]; !ok {
			out = append(out, name)
		}
	}
	return out
}

// Resolve substitutes every <var> placeholder in template with its bound value.
// It returns an error listing any unbound variables rather than emitting a command
// with empty holes — a half-filled command is worse than a clear error.
func (s *Store) Resolve(template string) (string, error) {
	if missing := s.Missing(template); len(missing) > 0 {
		return "", fmt.Errorf("vars: unresolved variables: %v", missing)
	}
	out := placeholderRe.ReplaceAllStringFunc(template, func(match string) string {
		name := match[1 : len(match)-1] // strip < >
		return s.values[name]
	})
	return out, nil
}

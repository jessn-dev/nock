// Package format defines nock's cheatsheet schema — the stable, versioned public
// contract between cheatsheet authors, the engine, and every frontend (TUI, MCP, AI).
//
// This is the project's long-term asset. Treat changes here as API changes:
// bump SchemaVersion and provide a migration. Fields beyond arsenal's original
// model (Intent, Tags, Risk, RequiresAuth) exist so the AI ranker (Milestone 4)
// has structured signal to rank on. They are optional and degrade gracefully.
package format

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// SchemaVersion is the current cheatsheet schema version. Cheatsheets declare the
// version they target; the loader rejects unknown majors and migrates known ones.
const SchemaVersion = "1"

// Risk classifies how destructive or noisy a command is. Frontends surface it;
// the AI ranker can de-prioritise high-risk commands unless intent calls for them.
type Risk string

const (
	RiskInfo        Risk = "info"   // passive, read-only (e.g. nslookup)
	RiskLow         Risk = "low"    // active but benign (e.g. a port scan)
	RiskMedium      Risk = "medium" // intrusive (e.g. brute force, exploit attempt)
	RiskHigh        Risk = "high"   // destructive or highly detectable
	RiskUnspecified Risk = ""       // author did not classify
)

var validRisks = map[Risk]bool{
	RiskInfo: true, RiskLow: true, RiskMedium: true, RiskHigh: true, RiskUnspecified: true,
}

// placeholderRe matches <name> variable placeholders inside a command template,
// e.g. "nmap -sV <target>" references the variable "target".
var placeholderRe = regexp.MustCompile(`<([a-zA-Z_][a-zA-Z0-9_]*)>`)

// Cheatsheet is a named collection of commands, typically one file.
type Cheatsheet struct {
	SchemaVersion string    `json:"schema_version" yaml:"schema_version"`
	Name          string    `json:"name"           yaml:"name"`
	Description   string    `json:"description"    yaml:"description,omitempty"`
	Author        string    `json:"author,omitempty" yaml:"author,omitempty"`
	Source        string    `json:"source,omitempty" yaml:"source,omitempty"` // provenance, e.g. imported-from URL
	License       string    `json:"license,omitempty" yaml:"license,omitempty"`
	Commands      []Command `json:"commands"       yaml:"commands"`
}

// Command is a single launchable command template plus the metadata the engine,
// the variable resolver, and the AI ranker need.
type Command struct {
	ID           string   `json:"id"            yaml:"id"`
	Name         string   `json:"name"          yaml:"name"`
	Command      string   `json:"command"       yaml:"command"` // template with <var> placeholders
	Description  string   `json:"description,omitempty" yaml:"description,omitempty"`
	Intent       string   `json:"intent,omitempty" yaml:"intent,omitempty"` // natural-language goal, for AI ranking
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Risk         Risk     `json:"risk,omitempty" yaml:"risk,omitempty"`
	RequiresAuth bool     `json:"requires_auth,omitempty" yaml:"requires_auth,omitempty"`
}

// Vars returns the distinct variable names referenced by the command template,
// in first-seen order. Used to prompt the operator for missing values.
func (c Command) Vars() []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range placeholderRe.FindAllStringSubmatch(c.Command, -1) {
		name := m[1]
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

// ErrValidation is returned (wrapped) when a cheatsheet or command is malformed.
var ErrValidation = errors.New("format: validation error")

// Validate checks a cheatsheet for structural integrity. It returns a joined error
// describing every problem found, not just the first, so authors fix in one pass.
func (cs Cheatsheet) Validate() error {
	var errs []error
	if strings.TrimSpace(cs.Name) == "" {
		errs = append(errs, fmt.Errorf("%w: cheatsheet name is required", ErrValidation))
	}
	if cs.SchemaVersion != "" && cs.SchemaVersion != SchemaVersion {
		errs = append(errs, fmt.Errorf("%w: unsupported schema_version %q (want %q)",
			ErrValidation, cs.SchemaVersion, SchemaVersion))
	}
	if len(cs.Commands) == 0 {
		errs = append(errs, fmt.Errorf("%w: cheatsheet %q has no commands", ErrValidation, cs.Name))
	}
	ids := map[string]bool{}
	for i, cmd := range cs.Commands {
		where := fmt.Sprintf("command[%d]", i)
		if strings.TrimSpace(cmd.ID) == "" {
			errs = append(errs, fmt.Errorf("%w: %s: id is required", ErrValidation, where))
		} else if ids[cmd.ID] {
			errs = append(errs, fmt.Errorf("%w: %s: duplicate id %q", ErrValidation, where, cmd.ID))
		} else {
			ids[cmd.ID] = true
		}
		if strings.TrimSpace(cmd.Command) == "" {
			errs = append(errs, fmt.Errorf("%w: %s (%s): command template is required",
				ErrValidation, where, cmd.ID))
		}
		if !validRisks[cmd.Risk] {
			errs = append(errs, fmt.Errorf("%w: %s (%s): invalid risk %q",
				ErrValidation, where, cmd.ID, cmd.Risk))
		}
	}
	return errors.Join(errs...)
}

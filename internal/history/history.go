// Package history records fired commands so an operator can recall and re-fire
// them mid-engagement. It is a frontend concern, not engine logic: the engine
// still owns search and resolution. History only persists what was fired.
//
// Security — secrets at rest: a resolved command can carry passwords, tokens, or
// target addresses. History stores the command *template* and the variable
// *bindings* separately, never a single flattened resolved string. That keeps a
// value redactable and means recall re-resolves through the engine rather than
// trusting a baked-in line. The file is created owner-only on every platform —
// POSIX 0600 on Unix, an owner-restricted DACL on Windows (see openAppend, which
// has per-OS implementations) — and NOCK_HISTORY=off disables persistence.
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry is one fired command. The template and the variable bindings are kept
// apart on purpose (see the package doc): the resolved string is never stored.
type Entry struct {
	Time     time.Time         `json:"time"`
	Sheet    string            `json:"sheet,omitempty"`
	ID       string            `json:"id"`
	Template string            `json:"template"`
	Vars     map[string]string `json:"vars,omitempty"`
}

// Store appends and reads history entries at a single file path. A Store with an
// empty path is a no-op sink (Append discards, Recent returns nothing); this is
// how NOCK_HISTORY=off and "no config dir" degrade gracefully.
type Store struct {
	path string
}

// New returns a Store writing to path. An empty path makes the Store a no-op.
func New(path string) *Store { return &Store{path: path} }

// Path reports the file the Store writes to ("" when disabled).
func (s *Store) Path() string { return s.path }

// DefaultPath resolves the history file location. It honours NOCK_HISTORY:
// "off" (case-insensitive) disables history and returns ""; any other value is
// used verbatim as the path. Otherwise it is <os.UserConfigDir>/nock/history.jsonl.
func DefaultPath() (string, error) {
	if v, ok := os.LookupEnv("NOCK_HISTORY"); ok {
		if isOff(v) {
			return "", nil
		}
		return v, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("history: locate config dir: %w", err)
	}
	return filepath.Join(dir, "nock", "history.jsonl"), nil
}

func isOff(v string) bool {
	for _, off := range []string{"off", "0", "false", "none"} {
		if strings.EqualFold(v, off) {
			return true
		}
	}
	return false
}

// Append writes one entry as a JSON line. It is a no-op for a disabled Store. The
// directory is created 0700 and the file is created owner-only (openAppend has a
// per-OS implementation) so secrets in variable bindings are not readable by
// other users on a shared host.
func (s *Store) Append(e Entry) error {
	if s.path == "" {
		return nil
	}
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("history: create dir: %w", err)
	}
	f, err := openAppend(s.path)
	if err != nil {
		return err
	}
	line, err := json.Marshal(e)
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("history: marshal: %w", err)
	}
	if _, err := fmt.Fprintf(f, "%s\n", line); err != nil {
		_ = f.Close()
		return fmt.Errorf("history: write: %w", err)
	}
	// Close error is reported: it can signal the record did not reach disk.
	if err := f.Close(); err != nil {
		return fmt.Errorf("history: close: %w", err)
	}
	return nil
}

// Recent returns up to n most-recent entries, newest first. A missing file (or a
// disabled Store) yields an empty slice, not an error. Malformed lines are
// skipped so one bad record never breaks recall.
func (s *Store) Recent(n int) ([]Entry, error) {
	if s.path == "" || n <= 0 {
		return nil, nil
	}
	f, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("history: open: %w", err)
	}
	defer func() { _ = f.Close() }() // read-only: a close error cannot lose data

	var all []Entry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // skip a corrupt record rather than fail the whole recall
		}
		all = append(all, e)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("history: read: %w", err)
	}

	// Newest first, capped at n.
	out := make([]Entry, 0, n)
	for i := len(all) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, all[i])
	}
	return out, nil
}

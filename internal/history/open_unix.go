//go:build !windows

package history

import (
	"fmt"
	"os"
)

// openAppend opens the history file for appending, creating it 0600 if absent.
// The kernel enforces owner-only access, so secrets in variable bindings are not
// readable by other users on the host. An *existing* file is re-tightened to 0600
// as well, so a file created or migrated with looser permissions cannot keep them
// while secrets are appended.
func openAppend(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) //nolint:gosec // G304: history path is operator-supplied by design
	if err != nil {
		return nil, fmt.Errorf("history: open: %w", err)
	}
	// O_CREATE only sets the mode on creation; enforce owner-only on every open so
	// a pre-existing, looser file is hardened before it receives secrets.
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("history: chmod: %w", err)
	}
	return f, nil
}

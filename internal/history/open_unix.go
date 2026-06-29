//go:build !windows

package history

import (
	"fmt"
	"os"
)

// openAppend opens the history file for appending, creating it 0600 if absent.
// The kernel enforces owner-only access, so secrets in variable bindings are not
// readable by other users on the host.
func openAppend(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) //nolint:gosec // G304: history path is operator-supplied by design
	if err != nil {
		return nil, fmt.Errorf("history: open: %w", err)
	}
	return f, nil
}

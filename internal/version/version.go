// Package version exposes build metadata stamped in at link time via -ldflags.
// See the Makefile and .goreleaser.yaml for the -X assignments.
package version

import (
	"fmt"
	"runtime"
)

// These are overridden at build time with:
//
//	go build -ldflags "-X github.com/jessn-dev/nock/internal/version.Version=v1.2.3 ..."
var (
	// Version is the semantic version, e.g. "v1.0.0". "dev" for unstamped builds.
	Version = "dev"
	// Commit is the short git SHA.
	Commit = "none"
	// Date is the build timestamp (RFC3339).
	Date = "unknown"
)

// String returns a one-line human-readable version banner.
func String() string {
	return fmt.Sprintf("nock %s (commit %s, built %s, %s/%s, %s)",
		Version, Commit, Date, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

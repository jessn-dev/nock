#!/usr/bin/env bash
#
# dev-check.sh — run nock's full local gate, mirroring CI as closely as a dev
# machine can. Beyond `make check` (which only exercises your host OS), this
# cross-builds and cross-lints the other release targets, so platform-specific
# breakage — e.g. a build-tagged Windows file that no longer compiles — is caught
# before you push, not on the CI runner.
#
# What it cannot do locally: run the test suite on other operating systems (that
# needs their runners). Cross-build + cross-vet + cross-lint catch compile and
# static issues; genuinely OS-specific runtime behaviour is still verified by CI.
#
# Usage:
#   scripts/dev-check.sh            # full gate
#   FAST=1 scripts/dev-check.sh     # skip the cross-platform build/lint matrix
#
# Exit non-zero on the first failing step.

set -euo pipefail

cd "$(dirname "$0")/.."

GO=${GO:-go}

# --- pretty output -----------------------------------------------------------
if [ -t 1 ]; then
	BLUE=$'\033[36m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; RED=$'\033[31m'; RESET=$'\033[0m'
else
	BLUE=''; GREEN=''; YELLOW=''; RED=''; RESET=''
fi

step()  { printf '\n%s==> %s%s\n' "$BLUE" "$1" "$RESET"; }
ok()    { printf '%s    ok%s\n' "$GREEN" "$RESET"; }
warn()  { printf '%s    skip: %s%s\n' "$YELLOW" "$1" "$RESET"; }
fail()  { printf '%s    FAIL: %s%s\n' "$RED" "$1" "$RESET"; exit 1; }

have()  { command -v "$1" >/dev/null 2>&1; }

# Cross-platform targets matching the release matrix.
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"

# --- 1. modules --------------------------------------------------------------
step "go mod tidy + verify (working tree must stay clean)"
cp go.mod go.mod.bak; cp go.sum go.sum.bak
$GO mod tidy
if ! diff -q go.mod go.mod.bak >/dev/null || ! diff -q go.sum go.sum.bak >/dev/null; then
	mv go.mod.bak go.mod; mv go.sum.bak go.sum
	fail "go.mod/go.sum not tidy — run 'go mod tidy' and commit the result"
fi
rm -f go.mod.bak go.sum.bak
$GO mod verify >/dev/null
ok

# --- 2. formatting (non-mutating check) --------------------------------------
step "gofmt (check only)"
unformatted=$(gofmt -l . 2>/dev/null || true)
if [ -n "$unformatted" ]; then
	printf '%s\n' "$unformatted"
	fail "files need gofmt — run 'make fmt'"
fi
ok

# --- 3. vet (host + cross-OS) ------------------------------------------------
step "go vet (host)"
$GO vet ./...
ok
if [ "${FAST:-0}" != "1" ]; then
	for p in $PLATFORMS; do
		os=${p%/*}; arch=${p#*/}
		step "go vet ($os/$arch)"
		GOOS=$os GOARCH=$arch $GO vet ./... && ok
	done
fi

# --- 4. lint (host + every target OS's build-tagged files) -------------------
# golangci-lint is mandatory here: a "full gate" that silently skips linting would
# green-light lint-broken changes locally. Fail if it is missing.
if ! have golangci-lint; then
	fail "golangci-lint not installed — required for 'make verify' (https://golangci-lint.run)"
fi
step "golangci-lint (host)"
golangci-lint run ./...
ok
if [ "${FAST:-0}" != "1" ]; then
	# Lint each target OS so OS-tagged files (e.g. *_windows.go) are covered, not
	# just the host's. GOARCH rarely changes which files compile, so one run per
	# unique GOOS is enough.
	host_os=$($GO env GOOS)
	# shellcheck disable=SC2086 # word-splitting PLATFORMS is intentional
	for os in $(printf '%s\n' $PLATFORMS | cut -d/ -f1 | sort -u); do
		[ "$os" = "$host_os" ] && continue
		step "golangci-lint (GOOS=$os)"
		GOOS=$os golangci-lint run ./... && ok
	done
fi

# --- 5. cross-build matrix ---------------------------------------------------
if [ "${FAST:-0}" != "1" ]; then
	for p in $PLATFORMS; do
		os=${p%/*}; arch=${p#*/}
		step "build ($os/$arch, CGO_ENABLED=0)"
		CGO_ENABLED=0 GOOS=$os GOARCH=$arch $GO build ./... && ok
	done
fi

# --- 6. tests (race + coverage) ----------------------------------------------
step "go test (race + coverage)"
$GO test -race -covermode=atomic -coverprofile=coverage.txt ./...
$GO tool cover -func=coverage.txt | tail -1
ok

# --- 7. vulnerabilities ------------------------------------------------------
step "govulncheck"
if have govulncheck; then
	govulncheck ./... && ok
else
	$GO run golang.org/x/vuln/cmd/govulncheck@latest ./... && ok
fi

# --- 8. secrets (optional) ---------------------------------------------------
step "gitleaks (optional)"
if have gitleaks; then
	gitleaks detect --no-banner --redact && ok
else
	warn "gitleaks not installed — CI still scans on the PR"
fi

printf '\n%sAll checks passed.%s\n' "$GREEN" "$RESET"

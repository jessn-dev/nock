# Contributing to nock

Thanks for your interest in nock. This guide gets you productive fast.

## Prerequisites

- Go (see `go.mod` for the minimum version)
- `make`
- Optional: `golangci-lint`, `goreleaser`, a local `ollama` for the offline AI path

## Development loop

```bash
make build      # compile to bin/nock
make test       # unit tests with the race detector
make check      # quick local gate: fmt, vet, lint, test (host OS only)
make verify     # full CI mirror: cross-platform build + lint, coverage, govulncheck
make run ARGS="search web directories"
```

`make check` is the fast pre-commit gate but only exercises your host OS. Before
pushing anything that touches platform-specific code (build-tagged files,
filesystem permissions, `os/exec`), run **`make verify`** — it cross-builds and
cross-lints every release target (`linux`, `darwin`, `windows`; `amd64`/`arm64`)
plus runs coverage and `govulncheck`, catching build-tag and compile breakage the
host-only gate misses. It cannot run another OS's *test suite* (CI does that), but
it catches the compile/lint failures that otherwise only surface on the runner.
The script is [`scripts/dev-check.sh`](scripts/dev-check.sh); `FAST=1` skips the
cross-platform matrix.

The scaffold is stdlib-only and builds offline. Heavy dependencies (bubbletea,
the MCP SDK, the Anthropic SDK, the YAML parser) are introduced milestone by
milestone — see `ROADMAP.md`.

## Architecture rules (non-negotiable)

1. **Engine-first.** All command search and variable resolution lives in
   `internal/engine`. Frontends (`internal/tui`, `internal/mcp`, future web) are thin
   shells that call the engine API and must not duplicate its logic.
2. **AI is optional.** Anything in `internal/ai` must degrade gracefully: with no
   provider or no network, fuzzy search still works. Never make AI a hard dependency
   of a core path.
3. **`pkg/format` is a public contract.** Schema changes are API changes — bump the
   schema version, update `schema.json`, and provide migration.
4. **Import, never bundle.** Do not copy arsenal source code or commit arsenal
   cheatsheet content. The importer fetches content onto the user's machine; it is
   never shipped in this repo or the binary. See `NOTICE`.
5. **No secrets.** API keys come from the environment, never from files, logs, or
   commits.

## Commits & PRs

- Use [Conventional Commits](https://www.conventionalcommits.org): `feat:`, `fix:`,
  `docs:`, `test:`, `chore:`, `refactor:`.
- Keep PRs focused; add tests for behavior changes.
- Fill in the PR template checklist. CI must be green.

### Branching (git-flow)

Feature work branches off `develop`, never `main`: `feature -> develop`, and
`develop -> main` only at release. Open PRs against `develop`; the
`develop -> main` PR is the release cut.

### Automated review (CodeRabbit)

**Why, on a solo project:** nock is single-maintainer, so there is no second
human reviewer to catch what the author misses. nock also *launches commands* —
it has to be safe for public use — so a bug or a permissions gap is a security
issue, not just a defect. CodeRabbit is the cheap, always-on "second pair of
eyes" that fills that gap: it already caught a real secrets-at-rest hole (history
files created before a fix kept loose permissions). It is a reviewer, not a gate
— it never blocks a merge and the maintainer decides every change.

PRs into `develop` and `main` are reviewed automatically by
[CodeRabbit](https://coderabbit.ai), configured in [`.coderabbit.yaml`](.coderabbit.yaml)
(CHILL profile, scoped to those two base branches). To (re)trigger a pass after
pushing fixes, comment `@coderabbitai review` on the PR. Treat its findings as
review input: verify each against the code, fix the valid ones, and say why
you're skipping the rest.

## Licensing of contributions

By contributing, you agree your contributions are licensed under the project's
[Apache-2.0](LICENSE) license.

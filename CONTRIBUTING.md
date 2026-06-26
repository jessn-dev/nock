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
make check      # full local gate: fmt, vet, lint, test — run before pushing
make run ARGS="search web directories"
```

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

## Licensing of contributions

By contributing, you agree your contributions are licensed under the project's
[Apache-2.0](LICENSE) license.

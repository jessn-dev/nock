# nock

[![CI](https://github.com/jessn-dev/nock/actions/workflows/ci.yml/badge.svg)](https://github.com/jessn-dev/nock/actions/workflows/ci.yml)
[![CodeQL](https://github.com/jessn-dev/nock/actions/workflows/codeql.yml/badge.svg)](https://github.com/jessn-dev/nock/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jessn-dev/nock)](https://goreportcard.com/report/github.com/jessn-dev/nock)
[![Go Reference](https://pkg.go.dev/badge/github.com/jessn-dev/nock.svg)](https://pkg.go.dev/github.com/jessn-dev/nock)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jessn-dev/nock)](go.mod)
[![Release](https://img.shields.io/github/v/release/jessn-dev/nock?include_prereleases&sort=semver)](https://github.com/jessn-dev/nock/releases)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> Nock the command, draw, fire.

**nock** is an AI-aware command launcher for pentesters and operators. It holds the
hard-to-remember commands you need mid-engagement, fills in your variables (target,
domain, wordlist) once, and hands the finished command to your shell — or to an AI
agent over [MCP](https://modelcontextprotocol.io). Single Go binary, no runtime
dependencies, provider-agnostic AI, works online or fully offline.

Inspired by [Orange-Cyberdefense/arsenal](https://github.com/Orange-Cyberdefense/arsenal),
rebuilt from scratch: engine-first, agent-native, vendor-neutral. See
[`ROADMAP.md`](ROADMAP.md) for the full plan and design decisions.

> ⚠️ For **authorized** security testing only. You are responsible for having
> permission to run what you launch. See [`SECURITY.md`](SECURITY.md).

## Why nock

- **Single static binary** (Go) — instant startup, trivial distribution, no
  Python/runtime drift.
- **Engine-first** — one command engine; the TUI, the MCP server, and a future web UI
  are thin frontends over it.
- **Agent-native** — ships an MCP server so Claude Code, Cursor, Continue, and other
  hosts can query your command knowledge directly. In MCP mode the host pays its own
  model cost; nock makes no LLM calls of its own.
- **Provider-agnostic AI** — command suggestion behind one interface: Anthropic,
  OpenAI-compatible (OpenAI, Groq, Together, vLLM, LocalAI, …), or **Ollama** for a
  fully offline / air-gapped path.
- **Graceful degradation** — fuzzy search needs no AI and no network. AI ranking
  layers on only when configured.

## Status

Early. **Milestone 0 (scaffold)** is in place: the engine, fuzzy search, the variable
resolver, the cheatsheet schema, and a scriptable CLI work today. The TUI, MCP server,
AI ranker, and arsenal importer are stubbed and land per the roadmap.

## Quick start

```bash
# Build
make build

# Search the example cheatsheets (offline, no AI)
./bin/nock search web directories

# Resolve a command's variables
./bin/nock resolve nmap-service-scan --var target=10.0.0.5
# -> nmap -sV -sC -oA scans/10.0.0.5 10.0.0.5

# Version / help
./bin/nock version
./bin/nock help
```

Point nock at your own cheatsheets:

```bash
export NOCK_CHEATSHEETS=/path/to/your/cheatsheets
./bin/nock search smb
```

## Modes

| Command | Mode | Status |
| --- | --- | --- |
| `nock` | Interactive TUI | stub (M2) |
| `nock --mcp` | MCP server over stdio | stub (M3) |
| `nock search <query>` | Non-interactive search | ✅ |
| `nock resolve <id> --var k=v` | Fill & print a command | ✅ |
| `nock import <src>` | Import arsenal cheatsheets | stub (M0/1) |
| `nock serve` | Team HTTP/SSE server | stub (M5) |

## Cheatsheet format

Cheatsheets are data and the project's stable contract. The Go definition lives in
[`pkg/format`](pkg/format/format.go); a JSON Schema is at
[`pkg/format/schema.json`](pkg/format/schema.json). Commands use `<name>`
placeholders, resolved from the variable store:

```json
{
  "schema_version": "1",
  "name": "recon",
  "commands": [
    {
      "id": "nmap-service-scan",
      "name": "nmap service/version scan",
      "command": "nmap -sV -sC -oA scans/<target> <target>",
      "intent": "identify open ports and the service versions behind them",
      "tags": ["nmap", "scan"],
      "risk": "low"
    }
  ]
}
```

YAML is the primary authoring format; JSON is accepted for tooling and
interchange. The codec is chosen by file extension, and a sheet round-trips
through either into the same structs.

## Documentation

Versioned, in-repo docs live in [`docs/`](docs/):

- [Install](docs/install.md) — source, `go install`, release binaries.
- [CLI reference](docs/cli.md) — every mode and flag.
- [Cheatsheet schema](docs/schema.md) — the `pkg/format` authoring contract.
- [MCP setup](docs/mcp.md) — exposing nock to AI agents (Milestone 3).

Narrative and community content (tutorials, FAQ, operator workflows) lives in the
[Wiki](https://github.com/jessn-dev/nock/wiki).

## Architecture

```
cmd/nock          single binary, mode switch
internal/engine   THE product: search + variable resolution (one source of truth)
internal/search   zero-dependency fuzzy matcher
internal/vars     global variable store (set once, fill everywhere)
internal/cheatsheet  load + validate cheatsheets
internal/ai       provider-agnostic ranker (Anthropic / OpenAI-compat / Ollama)
internal/mcp      MCP server frontend
internal/tui      Bubble Tea frontend
pkg/format        cheatsheet schema — the stable public contract
```

Frontends call the engine and nothing else; AI is always an optional layer. See
[`CONTRIBUTING.md`](CONTRIBUTING.md) for the architecture rules.

## License

[Apache-2.0](LICENSE). nock is an independent, clean-room project: no arsenal source
code is copied, and arsenal's GPL-3.0 cheatsheet content is never bundled — the
importer fetches it onto your machine. See [`NOTICE`](NOTICE).

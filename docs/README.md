# nock documentation

Versioned, in-repo documentation for nock. These docs ship with the tag and are
PR-reviewed, so they track the released binary. Narrative/community content
(tutorials, FAQ, operator workflows) lives in the [GitHub Wiki](https://github.com/jessn-dev/nock/wiki).

## Contents

- [Install](install.md) — build from source, `go install`, release binaries.
- [CLI reference](cli.md) — every mode and flag.
- [Cheatsheet schema](schema.md) — the `pkg/format` contract for authoring command sheets.
- [MCP setup](mcp.md) — exposing nock to AI agents (Milestone 3).

## Status

nock is pre-v1. Each page marks which milestone a feature lands in. Implemented
today: the engine (`search`, `resolve`) and both cheatsheet codecs (JSON + YAML).
The TUI, MCP server, AI ranker, and team server are scaffolded but not yet built —
see the [roadmap](../ROADMAP.md).

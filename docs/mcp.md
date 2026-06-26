# MCP setup

> **Status: Milestone 3 — not yet implemented.** This page documents the intended
> design so the contract is clear. `nock --mcp` currently returns a not-implemented
> error.

nock exposes its engine as a [Model Context Protocol](https://modelcontextprotocol.io)
server so AI agents (Claude Code, Cursor, Continue, Cline, Windsurf, Zed, …) can
query your command knowledge directly. MCP is provider-agnostic by protocol: one
server, every host, no per-host code.

**The host agent pays its own model cost.** In MCP mode nock makes no LLM calls of
its own — zero token spend on nock's side.

## Transport

`nock --mcp` runs an MCP server over stdio. Hosts launch it as a subprocess.

## Tools (planned)

| Tool | Input | Output |
|---|---|---|
| `search_commands` | intent / query | ranked commands |
| `get_command` | id (+ vars) | resolved command |
| `list_cheatsheets` | — | loaded cheatsheets |
| `resolve_vars` | target | populated variable set |

All tools are thin pass-throughs to the engine API. They are **read/suggest only**
— an agent can search and resolve, but nock never executes a command on the
agent's behalf. An AI response is untrusted network data; ranking never auto-runs.

## Example host config (planned)

Claude Code / Cursor style:

```json
{
  "mcpServers": {
    "nock": {
      "command": "nock",
      "args": ["--mcp"],
      "env": { "NOCK_CHEATSHEETS": "/path/to/your/cheatsheets" }
    }
  }
}
```

## Security

MCP requests are **untrusted input** (threat model: ADR 009). Tool arguments are
validated; path arguments are normalized and may not escape the configured
cheatsheet directory. The server exposes no write or exec tools. Secrets (API
keys) are never reachable through any tool output, log, or error.

# Cheatsheet schema

The cheatsheet format is nock's **stable public contract** (`pkg/format`). Treat
changes as API changes: the schema is versioned, and the engine, every frontend,
and the AI ranker all read these fields. The authoritative definitions are
[`pkg/format/format.go`](../pkg/format/format.go) and the JSON Schema in
[`pkg/format/schema.json`](../pkg/format/schema.json).

Current schema version: **`1`**.

## Authoring formats

YAML is the primary authoring format; JSON is accepted for tooling and
interchange. A sheet round-trips through either codec into the same structs — the
codec is chosen by file extension (`.yaml`/`.yml` or `.json`).

## Cheatsheet

A named collection of commands, typically one file.

| Field | Required | Description |
|---|---|---|
| `schema_version` | no | Targeted schema version. Empty is treated as current; an unknown version is rejected. |
| `name` | **yes** | Cheatsheet name. |
| `description` | no | Human description. |
| `author` | no | Author. |
| `source` | no | Provenance, e.g. the URL a sheet was imported from. |
| `license` | no | License of the sheet's content. |
| `commands` | **yes** | One or more commands (see below). |

## Command

A single launchable command template plus the metadata the engine, the variable
resolver, and the AI ranker need.

| Field | Required | Description |
|---|---|---|
| `id` | **yes** | Unique within the sheet. |
| `name` | no | Short display name. |
| `command` | **yes** | The command template, with `<var>` placeholders. |
| `description` | no | What the command does. |
| `intent` | no | Natural-language goal — the primary AI ranking signal (Milestone 4). |
| `tags` | no | Free-form tags; contribute to fuzzy search. |
| `risk` | no | One of `info` < `low` < `medium` < `high`. Empty = unspecified. |
| `requires_auth` | no | True if the command needs valid credentials/authorization. |

Fields beyond the original arsenal model (`intent`, `tags`, `risk`,
`requires_auth`) exist to give the AI ranker structured signal. They are optional
and degrade gracefully — fuzzy search ignores them when absent.

## Variables

Templates use `<name>` placeholders, e.g. `nmap -sV <target>`. Variable names
match `<[a-zA-Z_][a-zA-Z0-9_]*>`. Set a value once and it fills every template
that references it ("set once, fill everywhere"). A missing binding is an
**error**, never a half-filled command — see [the security model](#security).

## Validation

A sheet is rejected (with every problem reported at once) if:
- `name` is missing,
- `schema_version` is set to an unsupported value,
- it has no commands,
- any command is missing `id` or `command`,
- two commands share an `id`,
- a command's `risk` is not one of the known levels.

## Security

Cheatsheets are **untrusted, attacker-controlled data** — a malicious sheet is
nock's primary attack vector (threat model: ADR 009). nock never executes
cheatsheet content as its own code, and the operator always sees the fully
resolved command before it runs. What is displayed equals what runs; nock never
auto-fires.

## Example (YAML)

```yaml
schema_version: "1"
name: web
description: Web application probing starter commands.
author: nock
license: Apache-2.0
commands:
  - id: whatweb-fingerprint
    name: whatweb fingerprint
    command: whatweb -a 3 <url>
    intent: identify the web server, framework, and technologies in use
    tags: [web, fingerprint, enumeration]
    risk: low
```

See [`examples/cheatsheets/`](../examples/cheatsheets/) for working JSON and YAML sheets.

# Security Policy

nock is a tool for **authorized** security testing. Use it only against systems you
own or have explicit written permission to assess.

## Reporting a vulnerability

Do **not** open a public issue for security problems.

Report privately via GitHub Security Advisories:
<https://github.com/jessn-dev/nock/security/advisories/new>

Please include a description, reproduction steps, affected version (`nock version`),
and impact. We aim to acknowledge within 72 hours and to provide a remediation
timeline after triage. Coordinated disclosure is appreciated.

## Scope

In scope:
- The nock binary, engine, MCP server, and AI provider adapters.
- Handling of credentials/secrets (API keys must never be logged or persisted).
- Command resolution (a resolved command must never be silently mis-formed).

Out of scope:
- The behaviour of commands a user chooses to run — nock launches what the operator
  asks for. Operators are responsible for authorization and impact.
- Third-party cheatsheet content imported by the user.

## Handling of secrets

nock reads AI provider API keys from environment variables and never writes them to
disk, logs, or cheatsheet files. Report any deviation as a vulnerability.

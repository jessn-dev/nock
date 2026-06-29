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

## Automated security testing

Every pull request into `develop` and `main` runs a free, industry-standard
security stack in CI — the same baseline enterprises use at the free tier:

| Check | Tool | What it catches |
|---|---|---|
| SAST (Go) | [gosec](https://github.com/securego/gosec) via golangci-lint | injection, weak crypto, unsafe file perms |
| SAST (deep) | [CodeQL](https://codeql.github.com) (`security-and-quality`) | data-flow vulnerabilities |
| Dependency CVEs | [govulncheck](https://pkg.go.dev/golang.org/x/vuln) + Dependabot | known vulns in pinned deps |
| Secrets | [gitleaks](https://github.com/gitleaks/gitleaks) | API keys/tokens committed to history |
| Supply chain | [OpenSSF Scorecard](https://github.com/ossf/scorecard) | project security posture |

gosec and govulncheck gate every PR; CodeQL and gitleaks cover both merge hops
(`feature -> develop` and `develop -> main`). The binary is built `CGO_ENABLED=0`
with `-trimpath`; releases ship checksums and an SBOM.

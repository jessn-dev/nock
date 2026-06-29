# nock — Roadmap

> Nock the command, draw, fire. A fast, AI-aware command launcher for pentesters and operators.

`nock` is an arsenal-inspired command launcher. It holds the hard-to-remember commands you
need mid-engagement, fills in your variables (target IP, domain, wordlist) once, and hands the
finished command to your shell — or to an AI agent over MCP. Single Go binary, no runtime
dependencies, provider-agnostic AI, works online or fully offline.

Inspired by [Orange-Cyberdefense/arsenal](https://github.com/Orange-Cyberdefense/arsenal),
rebuilt from scratch: engine-first, agent-native, vendor-neutral.

---

## Why nock exists

Arsenal proved the idea — fast recall of pentest commands beats memorizing flags. But it is a
Python TUI, single-maintainer, stale since late 2024, and predates the agent era. `nock` keeps
the idea and modernizes the architecture:

- **Single static binary** (Go) — no Python/runtime drift, instant startup, easy distribution.
- **Engine-first** — the command engine is a library; TUI and MCP are thin frontends over one API.
- **Agent-native** — ships an MCP server so Claude Code, Cursor, Continue, etc. can query your
  command knowledge directly. Arsenal predates this entirely.
- **Provider-agnostic AI** — AI command suggestion behind a `Provider` interface. Claude, OpenAI,
  Gemini, Ollama (offline), or any OpenAI-compatible endpoint. No vendor lock.
- **Graceful degradation** — fuzzy search works with zero AI and zero network. AI ranking layers
  on top only when configured.

---

## Differentiators (vs arsenal)

| Pillar | arsenal | nock |
|---|---|---|
| Runtime | Python, dep-sensitive | Single Go static binary |
| AI command suggest | none | Provider-agnostic (cloud + local) |
| Agent integration (MCP) | none | First-class MCP server, any host |
| Architecture | TUI-coupled | Engine-first library + thin frontends |
| Offline AI | n/a | Ollama / local model adapter |
| Team / shared cheatsheets | manual git | roadmap: sync server |

---

## Architecture

Core principle: **library-first**. The engine knows nothing about TUI or MCP. Frontends import it.

```
/cmd
  /nock         → TUI app (default binary mode)
  ...           → mode switch: `nock` (TUI), `nock --mcp` (MCP), `nock serve` (team, later)
/internal
  /engine       → THE product: search, command resolution, variable injection, ranking
  /cheatsheet   → load/parse YAML + Markdown cheatsheets, registry, file watch
  /vars         → global variable store (set once, fill everywhere)
  /search       → fuzzy match (fzf-style) + AI rank hook
  /ai           → provider-agnostic ranker
    /providers  → anthropic, openai, gemini, ollama, openai-compat
  /mcp          → MCP server wrapper over engine
  /tui          → bubbletea UI
/pkg
  /format       → cheatsheet schema (public, stable contract — the real asset)
```

### Key boundaries

- `engine` exposes ONE clean API: `Search(intent) -> []Command`, `Resolve(cmd, vars) -> string`.
  TUI and MCP both call only this. No business logic duplicated in any frontend.
- `ai` is a **pluggable ranker**, not core. Search works offline (fuzzy); AI suggest layers on
  top when a provider is configured. No provider, no key, no network → still useful.
- `pkg/format` is the **stable contract**. Cheatsheets are data. Versioned schema. Protect it —
  it is the long-term asset and the migration path off any implementation detail.

### One binary, mode switch

- `nock` → TUI
- `nock --mcp` → stdio MCP server (agent-facing)
- `nock serve` → HTTP/SSE team server (later pillar)

### Two independent AI axes

1. **LLM provider** (engine → model): abstracted behind `Provider` interface. Swappable by config.
2. **MCP host** (agent → engine): provider-agnostic by protocol. One MCP server works with every
   MCP-capable host. No per-host code.

```go
type Provider interface {
    Suggest(ctx context.Context, intent string, candidates []Command) ([]Ranked, error)
    Name() string
}
```

---

## Stack

- Language: **Go** (single static binary, fast iteration, first-class MCP + Anthropic SDKs)
- TUI: `bubbletea` + `bubbles` + `lipgloss` (Charm)
- Fuzzy search: `sahilm/fuzzy` (or shell out to `fzf` when present)
- MCP: official `modelcontextprotocol/go-sdk`
- AI default: `anthropics/anthropic-sdk-go`, model `claude-opus-4-8` (quality) / sonnet (cheap rank)
- Cheatsheet parse: `goccy/go-yaml` + Markdown
- Config: provider/model/base_url/key-from-env — swap provider with zero rebuild

---

## Build order (de-risk early — each layer testable without the one above)

### Milestone 0 — Foundation
- [ ] Repo scaffold: `go.mod`, dir tree, CI, license
- [ ] `pkg/format` cheatsheet schema (YAML + Markdown), versioned
- [ ] `cheatsheet` loader + registry + golden test fixtures

### Milestone 1 — Engine (the product, headless)
- [ ] `vars` global variable store (set once, fill everywhere)
- [ ] `engine.Search(intent)` fuzzy + `engine.Resolve(cmd, vars)`
- [ ] Golden tests on resolution + var injection — zero UI dependency
- [ ] `nock` CLI: minimal non-interactive resolve (prove engine end-to-end)

### Milestone 2 — TUI (shippable v1)
- [ ] bubbletea list + search + var prompts
- [ ] **Show-before-fire (security, hard rule):** always display the fully-resolved command
      to the operator before any execution/prefill. What is displayed must equal what runs —
      no hidden expansion, no auto-execution. This is the primary mitigation for the
      malicious-cheatsheet injection vector and must be designed in now, not retrofitted.
- [x] tmux pane / prefill-into-shell output (prefill into the shell line, operator hits Enter —
      never auto-run on nock's behalf) — `--fire=stdout|tmux` + per-command `ctrl+t` override;
      tmux uses `send-keys -l` (no Enter), gated on `$TMUX` so it never fires where it can't work
- [x] command history — recall with `ctrl+r`; stores template + var bindings (never the
      flattened resolved string), owner-only on every OS (0600 / Windows owner DACL),
      `NOCK_HISTORY=off` disables
- [ ] Ship: GitHub Releases, Homebrew tap, `go install`, AUR
- [ ] **Launch-day: enable GitHub Discussions** (community + traction, feeds the funding
      story — sponsors/grants want a visible community). Categories: Announcements,
      Cheatsheets (share/request sheets — seeds the M5 team-sync content library and proves
      demand), Q&A, Ideas, Show & tell. Time it *with* the v1 launch, not before — an empty
      forum reads as a dead project. Pin a norms post: shared cheatsheets are untrusted,
      attacker-controlled input (ADR 009 threat model); nock shows the resolved command and
      never auto-runs — review every command before firing.
- [ ] **Docs, split by stability:**
  - **In-repo `/docs`** (versioned, PR-reviewed, ships with the tag) for anything that must
    track a release: `pkg/format` cheatsheet schema, CLI reference, MCP setup, install. Docs
    that define the contract live next to the code that enforces it, so they can't silently drift.
  - **GitHub Wiki** for narrative/community content (tutorials, FAQ, operator workflows,
    contributor notes) — low-stakes if it lags a release; a separate repo, no PR gate.

### Milestone 3 — MCP server ("install to AI tools")
- [ ] `nock --mcp` stdio server over engine
- [ ] Tools: `search_commands`, `get_command`, `list_cheatsheets`, `resolve_vars`
- [ ] One-line install configs for Claude Code / Cursor / Continue
- [ ] MCP host pays its own model cost — zero LLM spend on nock's side

### Milestone 4 — AI ranker (provider-agnostic, offline-capable)
- [ ] `Provider` interface + `ranker`
- [ ] **`ollama` adapter first-class + tested** (air-gapped/offline path is core, not optional)
- [ ] Adapters: `anthropic`, `openai-compat` (covers ~all cloud backends), then native `gemini` on demand
- [ ] Offline default-capable: fuzzy (no AI) + local Ollama both work with zero internet
- [ ] Cloud (Claude/OpenAI) = optional quality upgrade for users with internet + BYO key
- [ ] Docs: "point nock at your local Ollama" as a primary install flow
- [ ] Dev with Ollama for $0 token cost; cloud only for quality checks

### Milestone 5 — Team / shared cheatsheets (optional, post-validation)
- [ ] `nock serve` HTTP/SSE transport + auth
- [ ] Sync, versioned, org-shared command libraries
- [ ] Optional hosted service (only place nock could incur infra cost)

---

## Cost model

- **Build:** ~$0 infra. All OSS deps. Cost = developer time.
- **AI dev/testing:** pennies via cloud; **$0 via Ollama** local adapter.
- **Runtime:** users bring own API key (self-funded), or run offline via Ollama. In MCP mode the
  host agent pays its own model cost — **zero LLM spend on nock's side**.
- **Distribution:** GitHub Releases / Homebrew / AUR / `go install` — $0.
- **Only paid path:** optional hosted team-sync server (VPS ~$5–20/mo) — last pillar, opt-in.

---

## Non-goals (v1)

- Not a C2 / exploitation framework — nock launches *your* authorized commands, nothing more.
- Not building its own AI model — agents/providers supply intelligence; nock supplies curated
  command knowledge.
- No GUI/web app at launch — TUI + MCP first; web is a later frontend over the same engine.

## Decisions (settled)

- **Air-gapped: supported (superset).** Offline-capable is a first-class mode, not deferred.
  Ollama adapter ships and is tested in v1. Reasoning: catches gov/defense/banking users who
  ban cloud AI, while still working online. No language change — nock talks HTTP to a local
  Ollama process; Ollama runs the model, nock stays a thin Go client. Same HTTP code, cloud or
  local, different URL. Rust would only matter if nock embedded its own inference — it never will.
- **Cheatsheet format: hybrid.** nock has its own clean, versioned, AI-native schema (intent
  tags, semantic descriptions, ranking hints, risk/auth flags). Ship `nock import <arsenal-repo>`
  to convert arsenal YAML → nock format as seed content. Build on arsenal's *content*, not its
  *format*. Get their hundreds of commands AND a format that can hold our differentiator.
- **License: Apache-2.0** (permissive + patent grant, monetizable — hosted team server / pro tier).
  **HARD RULE: import, never bundle.** nock is a clean-slate Go rewrite — no arsenal code copied,
  so GPL never attaches to nock's code. Arsenal cheatsheet content stays GPL and is fetched by the
  *user* onto *their* machine via the importer; it is NEVER shipped inside nock's repo or binary.
  This keeps nock's code fully under Apache-2.0 and monetizable. (Not legal advice — confirm the
  GPL-content question with an IP lawyer before a paid launch.)
- **Ranking: staged.** v1 = fuzzy text match (instant, free, offline, no AI dependency). AI
  smart-ranking layers on in Milestone 4 as an upgrade, never a gate. Optional usage-history
  weighting considered later. Ship useful without AI first.

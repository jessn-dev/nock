# CLI reference

nock is a single binary with several modes. The engine is shared; each mode is a
thin frontend over it.

```
nock                 interactive TUI (default)            [Milestone 2]
nock --mcp           MCP server over stdio (for agents)   [Milestone 3]
nock search <query>  non-interactive search               [available]
nock resolve <id>    fill a command's variables and print [available]
nock serve           team HTTP/SSE server                 [Milestone 5]
nock import <src>    import arsenal cheatsheets as seed    [Milestone 0/1]
nock version         print build metadata                 [available]
```

## Environment

| Variable | Default | Meaning |
|---|---|---|
| `NOCK_CHEATSHEETS` | `examples/cheatsheets` | Directory scanned for `.json`/`.yaml`/`.yml` cheatsheets. |
| `NOCK_HISTORY` | `<config-dir>/nock/history.jsonl` | History file path. Set to an explicit path to relocate it, or `off` to disable history entirely. |

Load errors are non-fatal: a bad file is reported on stderr and the rest still load.

`<config-dir>` is `os.UserConfigDir()` — `~/.config` (Linux), `~/Library/Application
Support` (macOS), `%AppData%` (Windows).

## `nock` (interactive TUI)

The default mode: fuzzy-search, fill `<var>` placeholders, review the resolved
command, then fire it.

```
--fire=stdout|tmux   where a confirmed command is delivered (default: stdout)
```

- **stdout** (default, every platform): the command is printed once after the UI
  tears down, for the shell to capture — nock never runs it.
- **tmux**: the command is *prefilled* into the current tmux pane via
  `tmux send-keys -l` with **no** trailing Enter, so you still fire it yourself.
  Only offered inside a tmux session (`$TMUX` set); unavailable on Windows.

Keys: `ctrl+t` on the confirm screen overrides the target to tmux for one command;
`ctrl+r` opens history.

### History

Fired commands are recalled with `ctrl+r`. nock stores each command's **template
and variable bindings separately** — never the flattened resolved string — so a
recall re-resolves through the engine and values stay redactable. The history file
is created **owner-only on every OS**: POSIX `0600` on Linux/macOS, an owner-only
ACL on Windows. Set `NOCK_HISTORY=off` to disable persistence.

## `nock search <query>`

Fuzzy-search loaded commands and print matches, best first. Offline, instant, no
AI. Matches against name, intent, description, command, and tags.

```console
$ nock search nmap
nmap-service-scan        [recon] nmap -sV -sC -oA scans/<target> <target>
nmap-full-tcp            [recon] nmap -p- --min-rate 5000 -oA scans/<target>-allports <target>
```

## `nock resolve <id> [--var k=v ...]`

Look up a command by `id`, substitute its `<var>` placeholders from `--var`
bindings, and print the runnable command. An unbound variable is an **error**, not
a half-filled command.

```console
$ nock resolve whatweb-fingerprint --var url=http://10.0.0.5
whatweb -a 3 http://10.0.0.5

$ nock resolve nmap-service-scan
nock: engine: resolve "nmap-service-scan": vars: unresolved variables: [target]
```

`--var` may appear in any position and repeat. Both `--var k=v` and `--var=k=v`
forms work.

## `nock version`

Prints version, commit, and build date (stamped via ldflags at build time).

## Security

nock **shows, never auto-fires.** `resolve` prints the command to stdout; running
it is the operator's explicit act. The interactive TUI (Milestone 2) will always
display the fully resolved command before any prefill — what is displayed equals
what runs, with no hidden expansion. This is the primary mitigation for the
malicious-cheatsheet injection vector (ADR 009).

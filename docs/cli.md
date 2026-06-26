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

Load errors are non-fatal: a bad file is reported on stderr and the rest still load.

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

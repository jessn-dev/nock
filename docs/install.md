# Install

nock is a single static binary with no runtime dependencies — it builds for
linux/macOS/windows on amd64/arm64 with `CGO_ENABLED=0`.

> **Status:** building from source works today. Release binaries, Homebrew, and AUR
> land with the Milestone 2 ship step.

## From source

Requires Go (see [`go.mod`](../go.mod) for the minimum version).

```console
$ git clone https://github.com/jessn-dev/nock
$ cd nock
$ make build      # stamps version/commit/date via ldflags, outputs to bin/nock
$ ./bin/nock version
```

Or install straight to your `GOBIN`:

```console
$ go install github.com/jessn-dev/nock/cmd/nock@latest
```

## Release binaries (Milestone 2)

Prebuilt, checksummed archives will be attached to each
[GitHub Release](https://github.com/jessn-dev/nock/releases). Download the archive
for your platform, verify the checksum, extract, and put `nock` on your `PATH`.

## Homebrew (Milestone 2)

```console
$ brew install jessn-dev/tap/nock
```

## Arch (AUR) (Milestone 2)

```console
$ yay -S nock
```

## First run

nock loads cheatsheets from `$NOCK_CHEATSHEETS` (default `examples/cheatsheets`).
Point it at your own directory of `.yaml`/`.json` sheets:

```console
$ export NOCK_CHEATSHEETS=~/.config/nock/cheatsheets
$ nock search smb
```

See the [CLI reference](cli.md) for every mode and the
[cheatsheet schema](schema.md) for authoring your own command sheets.

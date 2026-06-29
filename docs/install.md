# Install

nock is a single static binary with no runtime dependencies — it builds for
linux/macOS/windows on amd64/arm64 with `CGO_ENABLED=0`.

> **Status:** building from source and `go install` work today. The release
> pipeline (GitHub Releases, Linux packages, and a Homebrew cask) is wired via
> GoReleaser and publishes automatically on the first version tag.

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

## Release binaries

Prebuilt, checksummed archives (with an SBOM) are attached to each
[GitHub Release](https://github.com/jessn-dev/nock/releases) for
linux/macOS/windows on amd64/arm64. Download the archive for your platform, verify
it against `checksums.txt`, extract, and put `nock` on your `PATH`:

```console
# Linux — verify before extracting
$ sha256sum -c checksums.txt --ignore-missing
$ tar xzf nock_*_linux_amd64.tar.gz
$ ./nock version
```

On **macOS** verify with `shasum -a 256 -c checksums.txt --ignore-missing`, then
`tar xzf`. On **Windows**, check the hash with
`CertUtil -hashfile nock_<ver>_windows_amd64.zip SHA256` and compare it against
`checksums.txt` before unzipping.

## Linux packages (.deb / .rpm / .apk)

Native packages for the common ecosystems are attached to each release — no extra
repo to add. Download the one for your distro and arch, then:

```console
# Debian / Ubuntu / Kali
$ sudo dpkg -i nock_*_linux_amd64.deb

# Fedora / RHEL / openSUSE
$ sudo rpm -i nock_*_linux_amd64.rpm

# Alpine
$ sudo apk add --allow-untrusted nock_*_linux_amd64.apk
```

> nock is a single static binary, so it also just runs on **any** Linux distro —
> if yours isn't listed, grab the `tar.gz` above or use `go install`.

## Homebrew (macOS / Linux)

```console
$ brew install --cask jessn-dev/tap/nock
```

The cask clears the macOS quarantine attribute on install, so nock runs without a
Gatekeeper prompt.

## Arch

An AUR package (`nock-bin`) is planned but **not yet published**: the AUR froze new
maintainer signups in June 2026 after a wave of malicious commits, so we can't
register it for now. In the meantime, Arch runs the static binary fine — use the
`tar.gz` from [Releases](https://github.com/jessn-dev/nock/releases) or `go install`
(both above). The AUR package lands once signups reopen.

## First run

nock loads cheatsheets from `$NOCK_CHEATSHEETS` (default `examples/cheatsheets`).
Point it at your own directory of `.yaml`/`.json` sheets:

```console
$ export NOCK_CHEATSHEETS=~/.config/nock/cheatsheets
$ nock search smb
```

See the [CLI reference](cli.md) for every mode and the
[cheatsheet schema](schema.md) for authoring your own command sheets.

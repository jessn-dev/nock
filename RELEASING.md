# Releasing nock

nock ships as a single static binary via [GoReleaser](https://goreleaser.com). A
release is **one command** — push a version tag and CI does the rest: cross-platform
binaries, Linux packages, checksums, an SBOM, auto-generated notes, and (when
configured) Homebrew.

Releases are cut from `main`, following the repo's git-flow: `feature -> develop`,
then `develop -> main` for the release.

---

## What is automated

Pushing a `v*` tag triggers [`.github/workflows/release.yml`](.github/workflows/release.yml),
which runs GoReleaser against [`.goreleaser.yaml`](.goreleaser.yaml) to produce:

- **Binaries** for linux/macOS/windows × amd64/arm64 (`CGO_ENABLED=0`, `-trimpath`,
  version/commit/date stamped via ldflags).
- **Archives** (`.tar.gz`, `.zip` on Windows) + `checksums.txt` + a CycloneDX **SBOM**.
- **Linux packages**: `.deb`, `.rpm`, `.apk` (Debian/Ubuntu/Kali, Fedora/RHEL, Alpine).
- **Homebrew cask** pushed to the tap (only if `HOMEBREW_TAP_TOKEN` is set).
- **Release notes**: a grouped changelog (Features/Fixes) built from Conventional
  Commits, wrapped in a templated install/verify header and footer. No manual editing.

> Release notes are your commit subjects. Write clear `feat:` / `fix:` messages;
> `chore`/`ci`/`docs`/`test` are filtered out automatically.

---

## One-time setup (optional — Homebrew only)

GitHub Releases, the Linux packages, and `go install` need **no** secrets. Only
Homebrew does:

1. Create an empty public repo `github.com/jessn-dev/homebrew-tap`.
2. Create a Personal Access Token with `repo` scope on that tap.
3. Add it as a repo secret named **`HOMEBREW_TAP_TOKEN`**
   (Settings → Secrets and variables → Actions).

AUR (`nock-bin`) is wired but **disabled** — the AUR froze new-account signups in
June 2026 after malicious commits. To re-enable once it reopens: register the
account, add an `AUR_KEY` secret (SSH private key), then uncomment the `aurs:` block
in `.goreleaser.yaml` and the `AUR_KEY` line in `release.yml`.

---

## Cut a release

```bash
# 1. Make sure develop is green and merged to main (git-flow release).
git checkout main && git pull
git merge --no-ff develop
git push origin main

# 2. Tag the release (semver, vX.Y.Z) and push the tag — this fires the pipeline.
git tag -a v0.1.0 -m "nock v0.1.0"
git push origin v0.1.0
```

That is the whole release. Watch the **Release** workflow in the Actions tab; when
it's green the GitHub Release is published with all artifacts and notes attached.

### Pre-releases

Tags with a pre-release suffix are flagged as prereleases automatically:

```bash
git tag -a v0.1.0-rc1 -m "nock v0.1.0-rc1"
git push origin v0.1.0-rc1
```

---

## Dry run (no publish)

Validate the config and build everything locally without releasing:

```bash
make release-check       # validate .goreleaser.yaml
make release-snapshot    # build dist/ for every target, publish nothing
```

`make verify` also runs `goreleaser check` as part of the full local gate.

---

## After the release — verify

- **GitHub → Releases**: archives, `checksums.txt`, `.deb`/`.rpm`/`.apk`, and SBOM
  are attached; the notes show grouped Features/Fixes plus the install block.
- `go install github.com/jessn-dev/nock/cmd/nock@v0.1.0` works.
- If Homebrew is set up: `brew install --cask jessn-dev/tap/nock`.
- Spot-check a binary: download, `sha256sum -c checksums.txt --ignore-missing`,
  then `nock version` shows the tag.

---

## Versioning

nock follows [Semantic Versioning](https://semver.org). Until `v1.0.0` the API
(including the [`pkg/format`](pkg/format) cheatsheet schema) may change between
minor versions; schema changes bump `schema_version` and ship a migration.

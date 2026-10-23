# Provenance and synchronization

This repository is built from three open source projects (all MIT). The Go module path stays `github.com/vernette/warpscout`
on purpose: it keeps imports identical to the sources, so their files can be compared and merged without rewriting.

## Sources

| Project | Role here | Recorded base |
| --- | --- | --- |
| https://github.com/vernette/warpscout | original engine, `internal/warpscout` (package `main` renamed to `warpscout`) | tag `v0.16.0`, commit `db4ac9ebae8d942191b8e8351f2c3a37a375bd66` |
| https://github.com/openwarpkit/warpscout-android | base of this repository: `core`, `mobileapi`, `mobile_backend.go`, `scan_i1.go`, `junk_i1.go` | imported unchanged as the first commit of this repository |
| https://github.com/niklzz/warpscout-tg | Telegram reachability check, `internal/warpscout/telegram.go` | `master`, ported by hand (see below) |

The first commit of this repository is the warpscout-android tree exactly as received. Everything after it is the work of this repository.

## What differs from the sources

- **vernette/warpscout.** After `v0.16.0`, `internal/warpscout/main.go` was aligned with upstream master: `isMihomo` for config-type
  checks, the `-best-by speed` notice and the documented TUI fallback. Apart from the package name, `main.go` matches upstream.
- **warpscout-android.** The Android app (`android/`), its CI/release workflows and design documents are removed. `core` gained
  `ScanSpec`/`ExpertOptions` (preset rules resolved on the server) and the Telegram options/result; `mobile_backend.go` passes them through.
- **warpscout-tg.** That fork is based on an older upstream, so it cannot be merged. The probe (`telegram.go`) is copied; the integration is
  rewritten against the current code: `endpointResult` / `foundMsg` carry the result, `lessByLossRTT` ranks reachability first,
  `filterByTelegram` implements "only with Telegram". The flags `-tg` / `-tg-only` and the CLI table columns are not ported.
- **New:** `cmd/warpscout-web`, `internal/webserver`, `deploy/`, `Dockerfile.web`, `docker-compose.yml`.

## Updating from vernette/warpscout

```sh
git remote add upstream https://github.com/vernette/warpscout.git
git fetch upstream --tags
git diff <recorded-tag> upstream/master -- '*.go'     # what changed upstream since the recorded base
```

Upstream keeps everything in `package main` at the repository root; here the same files live in `internal/warpscout` as
`package warpscout`, with `func Main()` instead of `func main()` in `cli.go`. Port changes file by file, keeping these local differences:
`mobile_backend.go`, `scan_i1.go`, `junk_i1.go`, `telegram.go`, the Telegram fields in `report.go`/`tui.go`/`main.go`,
and the `findjunk.go` refactor. Then run `go test ./...` and update the recorded tag and commit above.

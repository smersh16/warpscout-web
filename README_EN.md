# WarpScout Web

A web UI for finding working Cloudflare WARP endpoints. It runs on an Ubuntu server or in Docker and is driven from a browser.
Scanning happens on the server, where the WARP account and the scan history are stored. It can also check whether Telegram
works through a given endpoint.

[Русская версия](README.md)

> An independent project built on other authors' open source code (see [Sources](#sources-and-credits)).
> It is not an official release and is not affiliated with the authors of the original projects.

## Features

- Standard, Durable and Full presets; WireGuard, AmneziaWG, MASQUE H3 and MASQUE H2; IPv4/IPv6.
- Country and node filters (including exclusions), custom ranges, MTU, DNS, speed test, port sweep, WARP-in-WARP.
- AWG junk / I1 and MASQUE SNI finders; results are saved and can be applied to a scan with one click.
- **Telegram check:** a real MTProto request is sent to all 5 Telegram data centres through every tunnel. An endpoint counts only
  when all five answer, and those are ranked first. "Only with Telegram" keeps just them.
- Live results table, config export (WireGuard, AmneziaWG, usque, Mihomo), text report, CSV, copy-best-address.
- Scan history that restores every form value.
- SOCKS5 proxy to test one endpoint (listens on the server's loopback only).
- A scan keeps running on the server if you close the tab; progress and results are there when you return.
- Russian and English UI.

## Quick start

### Docker

```bash
git clone https://github.com/smersh16/warpscout-web warpscout-web && cd warpscout-web
echo "WARPSCOUT_TOKEN=$(openssl rand -hex 24)" > .env
docker compose up -d --build
```

Open `http://127.0.0.1:8787/` and enter the token from `.env`. Data lives in the `warpscout-data` volume.
The port is published on `127.0.0.1` only; for access from another machine use an SSH tunnel
(`ssh -L 8787:127.0.0.1:8787 user@server`) or an HTTPS reverse proxy.

### Ubuntu (systemd)

```bash
sudo ./deploy/ubuntu/install.sh
```

Installs Go, builds the binary and creates the `warpscout-web` service. Details, HTTPS and security settings:
[deploy/ubuntu/README.md](deploy/ubuntu/README.md) (Russian).

### Try it without a network or an account

```bash
go run ./cmd/warpscout-web -listen 127.0.0.1:8787 -no-auth -demo
```

Demo mode replaces the core with a simulation so you can look at the UI.

## Security

Listens on `127.0.0.1` by default; token authentication, server-side sessions valid for 7 days, per-IP lockout against token
guessing, token re-entry to export the account (it holds the private key), CSRF protection. Over plain HTTP the token travels
unencrypted, so put HTTPS in front for any remote access.

## Architecture

```
browser  ⇄  cmd/warpscout-web + internal/webserver   HTTP, SSE, UI, sessions, storage
               ⇄  mobileapi.Start / Cancel            the core's JSON protocol (same as the Android app's)
                     ⇄  core.Engine, core.ScanSpec     operations, presets, one operation at a time
                           ⇄  internal/warpscout        the WARPSCOUT scan engine + Telegram check
```

Development: `go test ./...`. How the code is derived and how to sync it: [UPSTREAM.md](UPSTREAM.md).

## Sources and credits

Built on open source code. All three sources are MIT licensed and their copyright notices are kept in [LICENSE](LICENSE)
and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

| Project | Author | What was taken | Where in this repository |
| --- | --- | --- | --- |
| [vernette/warpscout](https://github.com/vernette/warpscout) | Nikita S. (@vernette) | The original engine: WARP account registration, scanning, tunnels, junk/SNI finders, config generation, SOCKS | `internal/warpscout` (v0.16.0 base plus selected fixes from master), `main.go` |
| [openwarpkit/warpscout-android](https://github.com/openwarpkit/warpscout-android) | OpenWarpKit | The base of this repository: a multi-client wrapper around the engine (JSON protocol, `core`, `mobileapi`, mobile backend, automatic I1) | `core`, `mobileapi`, `internal/warpscout/mobile_backend.go`, `scan_i1.go`, `junk_i1.go` |
| [niklzz/warpscout-tg](https://github.com/niklzz/warpscout-tg) | niklzz | The idea and code of the Telegram reachability check (MTProto `req_pq_multi` to every DC) | `internal/warpscout/telegram.go` |

Also used: [amneziawg-go](https://github.com/amnezia-vpn/amneziawg-go), [usque](https://github.com/Diniboy1123/usque),
[WireGuard Go](https://github.com/WireGuard/wireguard-go), [gVisor](https://github.com/google/gvisor),
[Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss),
[quic-go](https://github.com/quic-go/quic-go) and other modules from [go.mod](go.mod); the full list with licenses is in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

### What changed relative to the sources

- Added: the web server and UI, server-side sessions and login protection, history and account storage, Docker and Ubuntu deployment.
- `core`: `ScanSpec` and `Resolve` (preset rules on the server, one source of truth), Telegram options and result.
- `internal/warpscout`: the Telegram check ported from warpscout-tg onto a newer base (that fork is built on an older version, so it was
  ported by hand), Telegram-aware result ordering, the "only with Telegram" filter, selected fixes from WARPSCOUT master.
- Not carried over: the Android app, the CLI flags `-tg` and `-tg-only` and the Telegram columns in the CLI tables.

The Android app this wrapper comes from lives in [openwarpkit/warpscout-android](https://github.com/openwarpkit/warpscout-android);
get console binaries of WARPSCOUT from the [original releases](https://github.com/vernette/warpscout/releases).

## Responsible use

This tool selects Cloudflare WARP endpoints. Use it in accordance with Cloudflare's terms and the laws of your country.

## License

MIT, see [LICENSE](LICENSE): the original project's copyright notice (Nikita S.) is kept and the owner of this repository is added; this repository's own additions are
distributed under the same terms.

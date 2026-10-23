#!/usr/bin/env bash
# Build and install WarpScout web on Ubuntu 22.04+ as a systemd service.
#   sudo ./deploy/ubuntu/install.sh
# Options (environment): LISTEN=127.0.0.1:8787 (default)  TOKEN=<fixed token>  TRUST_PROXY=1
#   LISTEN=0.0.0.0:8787 opens the service to the network (use HTTPS in front of it).
set -euo pipefail

[ "$(id -u)" -eq 0 ] || { echo "Run as root: sudo $0" >&2; exit 1; }
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
LISTEN="${LISTEN:-127.0.0.1:8787}"   # loopback by default; see README for HTTPS / remote access
GO_NEED="$(awk '/^go /{print $2; exit}' "$ROOT/go.mod")"

apt-get update -qq
apt-get install -y -qq ca-certificates curl tar >/dev/null

have_go=""
if command -v go >/dev/null 2>&1; then
  cur="$(go env GOVERSION | sed 's/^go//')"
  if [ "$(printf '%s\n%s\n' "$GO_NEED" "$cur" | sort -V | head -1)" = "$GO_NEED" ]; then have_go=1; fi
fi
if [ -z "$have_go" ]; then
  echo ">> Installing Go $GO_NEED to /usr/local/go-$GO_NEED"
  arch="$(dpkg --print-architecture)"   # amd64 / arm64
  curl -fsSL "https://go.dev/dl/go${GO_NEED}.linux-${arch}.tar.gz" -o /tmp/go.tgz
  rm -rf "/usr/local/go-$GO_NEED" && mkdir -p "/usr/local/go-$GO_NEED"
  tar -C "/usr/local/go-$GO_NEED" --strip-components=1 -xzf /tmp/go.tgz && rm /tmp/go.tgz
  export PATH="/usr/local/go-$GO_NEED/bin:$PATH"
fi

echo ">> Building"
cd "$ROOT"
VERSION="$(git describe --tags --always 2>/dev/null || echo dev)"
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /usr/local/bin/warpscout-web ./cmd/warpscout-web

id warpscout >/dev/null 2>&1 || useradd --system --home /var/lib/warpscout-web --shell /usr/sbin/nologin warpscout
install -d -o warpscout -g warpscout -m 700 /var/lib/warpscout-web

umask 077
{
  echo "WARPSCOUT_LISTEN=$LISTEN"
  [ -n "${TOKEN:-}" ] && echo "WARPSCOUT_TOKEN=$TOKEN"
  [ -n "${TRUST_PROXY:-}" ] && echo "WARPSCOUT_TRUST_PROXY=1"
  true
} > /etc/warpscout-web.env
chmod 640 /etc/warpscout-web.env; chown root:warpscout /etc/warpscout-web.env

install -m 644 "$ROOT/deploy/ubuntu/warpscout-web.service" /etc/systemd/system/warpscout-web.service
systemctl daemon-reload
systemctl enable --now warpscout-web
systemctl restart warpscout-web
sleep 1

port="${LISTEN##*:}"
echo
echo "WarpScout web is running on $LISTEN"
if [ -n "${TOKEN:-}" ]; then echo "Access token: $TOKEN"
else echo "Access token: $(cat /var/lib/warpscout-web/token 2>/dev/null || echo '(see journalctl -u warpscout-web)')"; fi
case "$LISTEN" in
  127.*|localhost:*|\[::1\]:*)
    echo "Listening on loopback only. From your computer:  ssh -L ${port}:127.0.0.1:${port} user@server  ->  http://127.0.0.1:${port}/"
    echo "For direct access: put HTTPS in front (deploy/ubuntu/Caddyfile.example) or re-run with LISTEN=0.0.0.0:${port} (plain HTTP!)." ;;
  *)
    echo "Open: http://<server-ip>:${port}/"
    echo "WARNING: plain HTTP sends the token unencrypted. Put it behind HTTPS (see deploy/ubuntu/README.md)." ;;
esac
echo "Logs: journalctl -u warpscout-web -f"

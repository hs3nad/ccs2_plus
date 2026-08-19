#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

BACKEND_SERVICE_SRC="$ROOT_DIR/boards/raspi1/systemd/hotel-maid-backend.service"
FRPC_SERVICE_SRC="$ROOT_DIR/boards/raspi1/systemd/frpc.service"
JOURNALD_SRC="$ROOT_DIR/boards/raspi1/systemd/journald.conf"

BACKEND_SERVICE_DST="/etc/systemd/system/hotel-maid-backend.service"
FRPC_SERVICE_DST="/etc/systemd/system/frpc.service"
JOURNALD_DST="/etc/systemd/journald.conf"

DO_RESTART=1

usage() {
  cat <<EOF
Usage:
  ./scripts/setup_raspi1.sh [--no-restart]

What it does:
  - installs systemd templates for hotel-maid-backend and frpc
  - installs journald.conf with Storage=volatile
  - reloads systemd
  - enables hotel-maid-backend and frpc
  - optionally restarts journald and both services
  - removes /var/log/journal to avoid persistent logs on storage

Options:
  --no-restart   install files and enable services without restarting them
  -h, --help     show this help
EOF
}

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    echo "Error: required file not found: $path" >&2
    exit 1
  fi
}

for arg in "$@"; do
  case "$arg" in
    --no-restart)
      DO_RESTART=0
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      usage >&2
      exit 1
      ;;
  esac
done

require_file "$BACKEND_SERVICE_SRC"
require_file "$FRPC_SERVICE_SRC"
require_file "$JOURNALD_SRC"

if [[ $EUID -ne 0 ]]; then
  echo "Please run as root: sudo ./scripts/setup_raspi1.sh" >&2
  exit 1
fi

echo "==> Installing systemd unit templates"
install -m 0644 "$BACKEND_SERVICE_SRC" "$BACKEND_SERVICE_DST"
install -m 0644 "$FRPC_SERVICE_SRC" "$FRPC_SERVICE_DST"

echo "==> Installing journald config"
install -m 0644 "$JOURNALD_SRC" "$JOURNALD_DST"

echo "==> Reloading systemd"
systemctl daemon-reload

echo "==> Enabling services"
systemctl enable hotel-maid-backend frpc

echo "==> Removing persistent journal directory if present"
rm -rf /var/log/journal

if [[ "$DO_RESTART" -eq 1 ]]; then
  echo "==> Restarting journald and services"
  systemctl restart systemd-journald
  systemctl restart hotel-maid-backend frpc
  echo "==> Done"
  echo "Check status with:"
  echo "  systemctl status hotel-maid-backend"
  echo "  systemctl status frpc"
else
  echo "==> Installed without restarting services"
  echo "Run manually when ready:"
  echo "  sudo systemctl restart systemd-journald"
  echo "  sudo systemctl restart hotel-maid-backend frpc"
fi

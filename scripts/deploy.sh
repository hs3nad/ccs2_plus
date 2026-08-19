#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BOARD="${BOARD:-lyra}"
TARGET="${1:-all}"

DIST_DIR="$ROOT_DIR/dist/$BOARD"

# ── Board config ────────────────────────────────────────────────────────────
case "$BOARD" in
  lyra)
    DEPLOY_METHOD="adb"
    ADB_SERIAL="${ADB_SERIAL:-}"
    REMOTE_BIN="${REMOTE_BIN:-/usr/local/bin/hotel-maid-backend}"
    REMOTE_DASHBOARD="${REMOTE_DASHBOARD:-/opt/hotel-dashboard}"
    REMOTE_INIT="${REMOTE_INIT:-/etc/init.d/S95hotel-maid-backend}"
    ;;
  raspi1)
    DEPLOY_METHOD="ssh"
    PI_HOST="${PI_HOST:-}"
    PI_USER="${PI_USER:-kaina}"
    REMOTE_BIN="${REMOTE_BIN:-/usr/local/bin/hotel-maid-backend}"
    REMOTE_DASHBOARD="${REMOTE_DASHBOARD:-/opt/hotel-dashboard}"
    REMOTE_SERVICE="${REMOTE_SERVICE:-hotel-maid-backend}"
    ;;
  *)
    echo "Unsupported BOARD: $BOARD" >&2
    echo "Supported: lyra, raspi1" >&2
    exit 1
    ;;
esac

# ── Helpers ──────────────────────────────────────────────────────────────────
adb_cmd() {
  if [[ -n "${ADB_SERIAL:-}" ]]; then
    adb -s "$ADB_SERIAL" "$@"
  else
    adb "$@"
  fi
}

ssh_cmd() { ssh "${PI_USER}@${PI_HOST}" "$@"; }

require_pi_host() {
  if [[ -z "${PI_HOST:-}" ]]; then
    echo "Error: PI_HOST must be set for $BOARD" >&2
    echo "  PI_HOST=192.168.x.x ./scripts/deploy_raspi1.sh $TARGET" >&2
    exit 1
  fi
}

require_dist() {
  local path="$1"
  if [[ ! -e "$path" ]]; then
    echo "Error: $path not found — build first:" >&2
    echo "  ./scripts/build_${BOARD}.sh" >&2
    exit 1
  fi
}

# ── Deploy steps ──────────────────────────────────────────────────────────────
_deploy_backend_adb() {
  echo "==> [lyra] push backend"
  require_dist "$DIST_DIR/go-backend"
  adb_cmd push "$DIST_DIR/go-backend" "$REMOTE_BIN"
  adb_cmd shell chmod +x "$REMOTE_BIN"
}

_deploy_dashboard_adb() {
  if [[ ! -d "$DIST_DIR/dashboard-web" ]]; then
    echo "    dashboard-web not found in dist — skipping"
    return
  fi
  echo "==> [lyra] push dashboard-web"
  adb_cmd shell mkdir -p "$REMOTE_DASHBOARD"
  adb_cmd push "$DIST_DIR/dashboard-web/." "$REMOTE_DASHBOARD/"
}

_restart_adb() {
  echo "==> [lyra] restart service"
  adb_cmd shell "$REMOTE_INIT restart"
}

_deploy_backend_ssh() {
  echo "==> [raspi1] rsync backend → $PI_USER@$PI_HOST"
  require_dist "$DIST_DIR/go-backend"
  rsync -avz --progress "$DIST_DIR/go-backend" "${PI_USER}@${PI_HOST}:/tmp/hotel-maid-backend-new"
  ssh_cmd "sudo mv /tmp/hotel-maid-backend-new $REMOTE_BIN && sudo chmod +x $REMOTE_BIN"
}

_deploy_dashboard_ssh() {
  if [[ ! -d "$DIST_DIR/dashboard-web" ]]; then
    echo "    dashboard-web not found in dist — skipping"
    return
  fi
  echo "==> [raspi1] rsync dashboard-web → $PI_USER@$PI_HOST"
  ssh_cmd "sudo mkdir -p $REMOTE_DASHBOARD && sudo chown $PI_USER $REMOTE_DASHBOARD"
  rsync -avz --delete --progress "$DIST_DIR/dashboard-web/" "${PI_USER}@${PI_HOST}:${REMOTE_DASHBOARD}/"
}

_restart_ssh() {
  echo "==> [raspi1] restart service"
  ssh_cmd "sudo systemctl restart $REMOTE_SERVICE"
}

# ── Usage ─────────────────────────────────────────────────────────────────────
usage() {
  cat <<EOF
Usage:
  ./scripts/deploy.sh [all|backend|dashboard|restart]

Environment:
  BOARD=lyra|raspi1          default: lyra
  ADB_SERIAL=<serial>        lyra: target specific ADB device (optional)
  PI_HOST=<ip>               raspi1: required
  PI_USER=<user>             raspi1: default kaina
  REMOTE_BIN=<path>          default /usr/local/bin/hotel-maid-backend
  REMOTE_DASHBOARD=<path>    default /opt/hotel-dashboard
  REMOTE_INIT=<path>         lyra: default /etc/init.d/S95hotel-maid-backend
  REMOTE_SERVICE=<name>      raspi1: default hotel-maid-backend

Examples:
  ./scripts/deploy_lyra.sh all
  PI_HOST=192.168.0.204 ./scripts/deploy_raspi1.sh all
  PI_HOST=192.168.0.204 ./scripts/deploy_raspi1.sh restart
EOF
}

# ── Main ─────────────────────────────────────────────────────────────────────
case "$TARGET" in
  all)
    if [[ "$DEPLOY_METHOD" == "adb" ]]; then
      _deploy_backend_adb
      _deploy_dashboard_adb
      _restart_adb
    else
      require_pi_host
      _deploy_backend_ssh
      _deploy_dashboard_ssh
      _restart_ssh
    fi
    ;;
  backend)
    if [[ "$DEPLOY_METHOD" == "adb" ]]; then
      _deploy_backend_adb
      _restart_adb
    else
      require_pi_host
      _deploy_backend_ssh
      _restart_ssh
    fi
    ;;
  dashboard)
    if [[ "$DEPLOY_METHOD" == "adb" ]]; then
      _deploy_dashboard_adb
      _restart_adb
    else
      require_pi_host
      _deploy_dashboard_ssh
      _restart_ssh
    fi
    ;;
  restart)
    if [[ "$DEPLOY_METHOD" == "adb" ]]; then
      _restart_adb
    else
      require_pi_host
      _restart_ssh
    fi
    ;;
  help | -h | --help)
    usage
    ;;
  *)
    echo "Unknown target: $TARGET" >&2
    usage >&2
    exit 1
    ;;
esac

echo "==> Done ($BOARD/$TARGET)"

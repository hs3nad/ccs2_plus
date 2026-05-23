#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BOARD="${BOARD:-lyra}"
TARGET="${1:-all}"

BACKEND_DIR="$ROOT_DIR/go-backend"
DASHBOARD_DIR="$ROOT_DIR/flutter-dashboard"
DIST_ROOT="$ROOT_DIR/dist"

case "$BOARD" in
  lyra)
    GOOS_TARGET="linux"
    GOARCH_TARGET="arm"
    GOARM_TARGET="7"
    DASHBOARD_PORT="6002"
    SSH_PORT="6102"
    ;;
  raspi1)
    GOOS_TARGET="linux"
    GOARCH_TARGET="arm"
    GOARM_TARGET="6"
    DASHBOARD_PORT="6010"
    SSH_PORT="6110"
    ;;
  *)
    echo "Unsupported BOARD: $BOARD" >&2
    echo "Supported boards: lyra, raspi1" >&2
    exit 1
    ;;
esac

DIST_DIR="$DIST_ROOT/$BOARD"
BACKEND_BIN="$DIST_DIR/go-backend"
DASHBOARD_OUT="$DIST_DIR/dashboard-web"
GOCACHE_DIR="${GOCACHE:-$ROOT_DIR/.cache/go-build}"

usage() {
  cat <<EOF
Usage:
  ./scripts/build.sh [all|backend|dashboard|clean]

Environment:
  BOARD=lyra|raspi1  default: lyra

Examples:
  ./scripts/build.sh all
  BOARD=raspi1 ./scripts/build.sh all
  ./scripts/build_lyra.sh backend
  ./scripts/build_raspi1.sh dashboard
EOF
}

prepare_dist() {
  mkdir -p "$DIST_DIR" "$GOCACHE_DIR"
}

write_manifest() {
  cat >"$DIST_DIR/build-info.txt" <<EOF
board=$BOARD
dashboard_port=$DASHBOARD_PORT
ssh_port=$SSH_PORT
goos=$GOOS_TARGET
goarch=$GOARCH_TARGET
goarm=$GOARM_TARGET
backend_binary=go-backend
dashboard_dir=dashboard-web
EOF
}

build_backend() {
  prepare_dist
  echo "==> Building backend for $BOARD ($GOOS_TARGET/$GOARCH_TARGET GOARM=$GOARM_TARGET)"
  (
    cd "$BACKEND_DIR"
    env \
      CGO_ENABLED=0 \
      GOOS="$GOOS_TARGET" \
      GOARCH="$GOARCH_TARGET" \
      GOARM="$GOARM_TARGET" \
      GOCACHE="$GOCACHE_DIR" \
      go build -buildvcs=false -trimpath -ldflags="-s -w" -o "$BACKEND_BIN" .
  )
}

build_dashboard() {
  prepare_dist
  echo "==> Building dashboard web for $BOARD (same-origin backend)"
  (
    cd "$DASHBOARD_DIR"
    flutter build web --release --dart-define=BACKEND_BASE_URL=
  )

  rm -rf "$DASHBOARD_OUT"
  mkdir -p "$DASHBOARD_OUT"
  cp -R "$DASHBOARD_DIR/build/web/." "$DASHBOARD_OUT/"
}

clean_board() {
  echo "==> Cleaning $DIST_DIR"
  rm -rf "$DIST_DIR"
}

case "$TARGET" in
  all)
    build_backend
    build_dashboard
    write_manifest
    ;;
  backend)
    build_backend
    write_manifest
    ;;
  dashboard)
    build_dashboard
    write_manifest
    ;;
  clean)
    clean_board
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    echo "Unsupported build target: $TARGET" >&2
    usage >&2
    exit 1
    ;;
esac

echo "==> Build output: $DIST_DIR"

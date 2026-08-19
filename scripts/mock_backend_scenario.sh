#!/usr/bin/env bash
set -euo pipefail

SCENARIO="${1:-normal}"
LISTEN_ADDR="${LISTEN_ADDR:-:18080}"
DELAYED_RECOVER_AFTER="${DELAYED_RECOVER_AFTER:-8s}"

case "$SCENARIO" in
  normal|delayed-room-update|delayed-room-update-recover|sse-drop-after-full-state)
    ;;
  *)
    echo "unsupported scenario: $SCENARIO" >&2
    echo "supported: normal | delayed-room-update | delayed-room-update-recover | sse-drop-after-full-state" >&2
    exit 1
    ;;
esac

exec go run -C go-backend ./cmd/mock-scenario \
  --listen "$LISTEN_ADDR" \
  --scenario "$SCENARIO" \
  --delayed-recover-after "$DELAYED_RECOVER_AFTER"

#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BOARD=lyra exec "$SCRIPT_DIR/deploy.sh" "${1:-all}"

#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PORT="${PORT:-9017}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${PORT}}"

PORT="${PORT}" "${ROOT_DIR}/scripts/restart_local.sh"
BASE_URL="${BASE_URL}" "${ROOT_DIR}/scripts/regression_inventory_smoke.py"

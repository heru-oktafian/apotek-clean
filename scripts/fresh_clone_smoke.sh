#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="${TMP_DIR:-/tmp/apotek-clean-fresh-smoke}"
PORT="${PORT:-9019}"
BASE_URL="http://127.0.0.1:${PORT}"
ENV_SOURCE="${ENV_SOURCE:-${ROOT_DIR}/.env}"
GO_BIN="${GO_BIN:-/usr/local/go/bin/go}"

cleanup() {
  if [[ -n "${APP_PID:-}" ]]; then
    kill "${APP_PID}" >/dev/null 2>&1 || true
    wait "${APP_PID}" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

rm -rf "${TMP_DIR}"

echo "[info] cloning repo ke ${TMP_DIR}..."
git clone "${ROOT_DIR}" "${TMP_DIR}" >/dev/null 2>&1

if [[ -f "${ENV_SOURCE}" ]]; then
  cp "${ENV_SOURCE}" "${TMP_DIR}/.env"
  echo "[info] .env disalin dari ${ENV_SOURCE}"
else
  echo "[warn] .env sumber tidak ditemukan di ${ENV_SOURCE}, lanjut tanpa copy env"
fi

cd "${TMP_DIR}"

echo "[info] build fresh clone..."
"${GO_BIN}" build -o ./bin/apotek ./cmd/app

echo "[info] menjalankan fresh clone di port ${PORT}..."
PORT="${PORT}" ./bin/apotek > "${TMP_DIR}/app.log" 2>&1 &
APP_PID=$!

echo "[info] menjalankan smoke regression terhadap fresh clone..."
BASE_URL="${BASE_URL}" ./scripts/regression_inventory_smoke.py

echo "[ok] fresh clone smoke lulus"

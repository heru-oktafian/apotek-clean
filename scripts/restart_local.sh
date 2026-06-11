#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PORT="${PORT:-9002}"
BINARY_PATH="${ROOT_DIR}/bin/apotek"
LOG_PATH="${ROOT_DIR}/app.log"
GO_BIN="${GO_BIN:-/usr/local/go/bin/go}"

kill_port() {
  local port="$1"

  if command -v lsof >/dev/null 2>&1; then
    local pids
    pids="$(lsof -ti tcp:"${port}" || true)"
    if [[ -n "${pids}" ]]; then
      echo "[info] menghentikan proses di port ${port}: ${pids}"
      kill ${pids} || true
      sleep 1
    fi
    return
  fi

  if command -v fuser >/dev/null 2>&1; then
    echo "[info] menghentikan proses di port ${port} via fuser"
    fuser -k "${port}/tcp" || true
    sleep 1
  fi
}

mkdir -p "${ROOT_DIR}/bin"
kill_port "${PORT}"

echo "[info] build binary..."
"${GO_BIN}" build -o "${BINARY_PATH}" "${ROOT_DIR}/cmd/app"

echo "[info] menjalankan app di port ${PORT}..."
nohup env PORT="${PORT}" "${BINARY_PATH}" > "${LOG_PATH}" 2>&1 &
PID=$!

echo "[ok] app hidup dengan pid ${PID}"
echo "[ok] log: ${LOG_PATH}"

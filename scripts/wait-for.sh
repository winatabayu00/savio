#!/usr/bin/env bash
# Wait for a TCP service to be reachable. Usage: wait-for.sh host port [timeout_seconds]
set -euo pipefail

HOST="${1:?host required}"
PORT="${2:?port required}"
TIMEOUT="${3:-60}"

for _ in $(seq 1 "$TIMEOUT"); do
  if nc -z "$HOST" "$PORT" 2>/dev/null; then
    echo "ready: ${HOST}:${PORT}"
    exit 0
  fi
  sleep 1
done

echo "timeout waiting for ${HOST}:${PORT}" >&2
exit 1
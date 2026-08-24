#!/usr/bin/env bash
# Savio one-command dev bootstrap: infra + DB + migrations + demo seed +
# backend + frontend. From a fresh clone: ./scripts/dev.sh, then open the URL.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

command -v docker >/dev/null || { echo "error: docker required (Docker Desktop on macOS)"; exit 1; }
command -v go >/dev/null || { echo "error: go required"; exit 1; }
command -v npm >/dev/null || { echo "error: node/npm required"; exit 1; }

echo "==> 1/6 environment"
test -f .env || { cp .env.example .env; echo "    .env created from .env.example"; }

echo "==> 2/6 infrastructure (postgres redis minio)"
docker compose up -d --wait postgres redis 2>/dev/null || docker compose up -d postgres redis
"$ROOT/scripts/wait-for.sh" localhost 5433 90
"$ROOT/scripts/wait-for.sh" localhost 6380 60

echo "==> 3/6 backend deps"
( cd backend && go mod download )

echo "==> 4/6 database schema + demo data"
set -a; . ./.env; set +a
( cd backend && go run ./cmd/migrate -action=up )
( cd backend && go run ./cmd/seed -action=demo )

echo "==> 5/6 frontend deps"
test -d frontend/node_modules || ( cd frontend && npm install )

echo "==> 6/6 starting servers"
( cd backend && exec go run ./cmd/api ) >/tmp/savio-api.log 2>&1 &
API_PID=$!
( cd frontend && exec npm run dev ) >/tmp/savio-web.log 2>&1 &
WEB_PID=$!
trap 'kill "$API_PID" "$WEB_PID" 2>/dev/null || true' EXIT INT TERM

"$ROOT/scripts/wait-for.sh" localhost 5173 120
echo
echo "Savio is running:"
echo "  Frontend  http://localhost:5173   (logs: /tmp/savio-web.log)"
echo "  API       http://localhost:8080   (logs: /tmp/savio-api.log)"
echo "  Demo user demo@savio.test / DemoPassword!23"
echo "Press Ctrl-C to stop."
wait
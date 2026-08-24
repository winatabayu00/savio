# Savio — Implementation Progress

This file tracks milestone completion. It is not a replacement for the
source-of-truth implementation plan (`docs/engineering/implementation-plan.md`).

## M00 — Repository & Documentation Bootstrap

Status: COMPLETED (pre-existing)

- Repository initialized with AGENTS.md, README.md, DESIGN.md
- Full documentation tree present (product, architecture, database, api, engineering, assignment)

## M01 — Infrastructure & Developer Experience

Status: COMPLETED

- docker-compose.yml (postgres :5433, redis :6380, minio :9000/:9001, minio-init)
- .env.example, Makefile, scripts/wait-for.sh, .github/workflows/ci.yml
- Verified: compose up, postgres healthy, redis PONG, minio bucket initialized

## M02 — Backend Foundation

Status: COMPLETED

- Go module, Gin router, platform packages (config, money, errs, httpx, authctx, database, redisclient, mw)
- /health, /ready (PostgreSQL + Redis), request ID, recovery without stack leak, CORS, security headers, graceful shutdown
- Verification: `go test ./...` PASS, `/health`/`/ready` runtime verified

## M03 — Database Foundation & Migrations

Status: COMPLETED

- golang-migrate with embedded iofs source, explicit up/down migrations
- Schema: users, workspaces, workspace_memberships, auth_sessions, user_settings, accounts, categories
- System categories seed (idempotent, 21 categories)
- Workspace-id scoping decision documented in database-design.md §3.1
- Verification: up/rollback/up PASS, migration test PASS, seed idempotent

## M04 — Authentication & Session Security

Status: IN PROGRESS
# Savio — Implementation Progress

This file records implemented repository behavior as of the current codebase. It is not a replacement for the source-of-truth implementation plan (`docs/engineering/implementation-plan.md`).

## M00 — Repository & Documentation Bootstrap

Status: COMPLETED (pre-existing)

- Repository initialized with AGENTS.md, README.md, DESIGN.md
- Full documentation tree present (product, architecture, database, api, engineering, assignment)

## M01 — Infrastructure & Developer Experience

Status: COMPLETED

- docker-compose.yml (postgres :5433, redis :6380, minio :9000/:9001, minio-init)
- .env.example, Makefile, scripts/wait-for.sh, .github/workflows/ci.yml
- Local Compose provisions PostgreSQL, Redis, and MinIO; backend currently uses Redis for rate limiting, not MinIO or a queue.

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

## M04-M22 — Application Delivery

Status: COMPLETED

- Cookie authentication, CSRF, refresh-token rotation, sessions, user settings, and workspace RBAC.
- Accounts, categories, transactions, transfers, recurring activity, analytics, budgets, goals, deterministic forecast, and non-destructive scenarios.
- AI configuration, categorization, insight generation, Copilot conversations, audit logging, and Telegram recap ingestion.
- Explicit migrations `000001` through `000016`, each with up/down files.
- React pages for the core finance workflow.
- Polling worker for recurring auto-posting, notification sweeps, and Telegram long-polling.

## Deferred / Not Implemented

- Production Docker images and deployment configuration.
- Redis/Asynq queue processing.
- File upload or backend MinIO object-storage integration.
- Tailwind-first frontend implementation; current UI uses Duralux SCSS.

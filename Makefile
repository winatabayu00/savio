include .env
export

ENV_FILE := .env

.PHONY: help infra-up infra-down infra-logs ps migrate-up migrate-down migrate-force \
        dev-api dev-worker dev-web test test-backend test-frontend lint build seed-demo \
        telegram-webhook fresh-db audit

help:
	@echo "Savio make targets"
	@echo "  infra-up / infra-down    docker compose up/down"
	@echo "  migrate-up / migrate-down  run database migrations"
	@echo "  migrate-force            force a migration step (MIGRATE_VERSION=n)"
	@echo "  dev-api / dev-worker     run backend binaries in watch mode"
	@echo "  dev-web                  run frontend dev server"
	@echo "  test / test-backend / test-frontend  run tests"
	@echo "  seed-demo                seed demo finance data"
	@echo "  fresh-db                 drop + recreate schema + migrate + seed"
	@echo "  audit                    run full verification suite"

infra-up:
	@test -f $(ENV_FILE) || cp .env.example $(ENV_FILE)
	docker compose up -d postgres redis minio minio-init

infra-down:
	docker compose down

infra-logs:
	docker compose logs -f --tail=200

ps:
	docker compose ps

migrate-up:
	@test -f $(ENV_FILE) || (echo ".env missing; copy .env.example" && exit 1)
	cd backend && go run ./cmd/migrate -action=up

migrate-down:
	cd backend && go run ./cmd/migrate -action=down

migrate-force:
	cd backend && go run ./cmd/migrate -action=force -version=$(MIGRATE_VERSION)

dev-api:
	cd backend && go run ./cmd/api

dev-worker:
	cd backend && go run ./cmd/worker

dev-web:
	cd frontend && npm run dev

test-backend:
	cd backend && go test ./...

test-backend-race:
	cd backend && go test -race ./...

test-frontend:
	cd frontend && npm run test -- --run

test:
	make test-backend test-frontend

lint:
	cd backend && go vet ./...
	cd frontend && npm run typecheck

seed-demo:
	cd backend && go run ./cmd/seed -action=demo

telegram-webhook:
	@test "$$WORKSPACE" != "" || (echo "WORKSPACE=<workspace-id> required" && exit 1)
	cd backend && go run ./cmd/telegramwebhook -workspace=$(WORKSPACE) -url=$(WEBHOOK_URL)

fresh-db:
	cd backend && go run ./cmd/migrate -action=down-all
	cd backend && go run ./cmd/migrate -action=up
	cd backend && go run ./cmd/seed -action=demo

audit:
	cd backend && go build ./... && go vet ./... && go test ./...
	cd frontend && npm run typecheck && npm run test -- --run && npm run build
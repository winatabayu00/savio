# Local Development

This guide covers installation, environment configuration, migrations, local services, and verification. Run commands from the repository root unless stated otherwise.

## Prerequisites

- Docker with Docker Compose
- Go version declared in `backend/go.mod`
- Node.js with npm
- Make

## Fastest Setup

```bash
git clone <repository-url>
cd savio
./scripts/dev.sh
```

The script:

1. Copies `.env.example` to `.env` when needed.
2. Starts PostgreSQL and Redis.
3. Downloads backend dependencies.
4. Applies migrations and seeds demo data.
5. Installs frontend dependencies.
6. Starts the API and frontend.

Open [http://localhost:5173](http://localhost:5173).

```text
Email:    demo@savio.test
Password: DemoPassword!23
```

API logs are written to `/tmp/savio-api.log`; frontend logs to `/tmp/savio-web.log`. Press `Ctrl-C` to stop the application processes.

## Environment

Create local configuration:

```bash
cp .env.example .env
```

The template contains PostgreSQL, Redis, cookie-authentication, CSRF, AI, and MinIO settings. Development defaults use:

```text
Frontend:   http://localhost:5173
API:        http://localhost:8080
PostgreSQL: localhost:5433
Redis:      localhost:6380
MinIO:      localhost:9000
```

Keep real credentials out of Git. AI defaults to the mock provider and does not require an external API key.

## Manual Setup

Start infrastructure:

```bash
make infra-up
```

Install dependencies:

```bash
cd backend && go mod download
cd ../frontend && npm install
cd ..
```

Prepare the database:

```bash
make migrate-up
make seed-demo
```

Run each process in a separate terminal:

```bash
make dev-api
make dev-web
make dev-worker
```

The worker is only required for asynchronous jobs.

## Database Migrations

Apply all pending migrations:

```bash
make migrate-up
```

Roll back the latest migration:

```bash
make migrate-down
```

Rebuild and reseed the development database:

```bash
make fresh-db
```

Force a migration version only when repairing migration state:

```bash
make migrate-force MIGRATE_VERSION=<version>
```

Migration files live in `backend/migrations/`. Production schema changes use explicit migrations, not GORM `AutoMigrate`.

## Common Commands

```bash
make help
make ps
make infra-logs
make infra-down
make seed-demo
```

An optional repository CLI wraps the same Make targets:

```bash
./scripts/savio help
./scripts/savio dev
./scripts/savio migrate
./scripts/savio test
./scripts/savio audit
```

## Testing

Backend tests:

```bash
make test-backend
```

Backend race detector:

```bash
make test-backend-race
```

Frontend tests:

```bash
make test-frontend
```

Backend and frontend tests:

```bash
make test
```

Static verification:

```bash
make lint
```

Full build, test, typecheck, and frontend build verification:

```bash
make audit
```

## API Documentation

- API contract: [`docs/api/api-contract.md`](../api/api-contract.md)

## Troubleshooting

Check infrastructure status and logs:

```bash
make ps
make infra-logs
```

Recreate local infrastructure without deleting repository files:

```bash
make infra-down
make infra-up
```

If migration state is valid but demo data needs refreshing, run `make seed-demo`. Use `make fresh-db` only when discarding local development data is acceptable.

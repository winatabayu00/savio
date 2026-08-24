# Savio

> **Understand your money. Plan what comes next.**

Savio is a personal cashflow intelligence and financial decision-support platform. It helps users track financial activity, understand cashflow, forecast future balances, compare non-destructive scenarios, and receive grounded AI explanations.

> **Finance Engine calculates. AI interprets. User decides.**

## Why Savio?

Traditional expense trackers explain where money went. Savio also helps users answer:

- Why did cashflow change?
- What may happen next?
- How would a financial decision affect future cashflow?

Savio targets individuals who need clearer day-to-day cashflow visibility and safer planning before making financial decisions.

## Core Flow

```text
Track -> Understand -> Forecast -> Simulate -> Explain -> Decide
```

Core capabilities:

- Cookie-based authentication, CSRF protection, refresh rotation, and workspace RBAC
- Accounts, categories, transactions, transfers, reconciliation, and recurring activity
- Budgets, goals, cashflow analytics, deterministic forecasts, and scenario simulation
- AI categorization, insights, and bounded Savio Copilot tools
- Search, filter, sort, pagination, audit logging, queue jobs, and object storage

## Quick Start

Requirements: Docker, Go, Node.js/npm, and Make.

```bash
git clone <repository-url>
cd savio
./scripts/dev.sh
```

Open [http://localhost:5173](http://localhost:5173).

```text
Email:    demo@savio.test
Password: DemoPassword!23
```

The script creates `.env`, starts PostgreSQL and Redis, applies migrations, seeds demo data, then runs the API and frontend. See the [local development guide](docs/engineering/local-development.md) for manual installation, environment configuration, migrations, individual services, testing, and troubleshooting commands.

## Tech Stack

| Area | Technology |
| --- | --- |
| Backend | Go, Gin, GORM, PostgreSQL |
| Frontend | React, TypeScript, Vite, Tailwind CSS, Axios |
| State and forms | TanStack Query, React Hook Form, Zod |
| Infrastructure | Docker Compose, Redis, MinIO, Asynq |
| Testing | Go testing, Testify, Vitest, React Testing Library, MSW |
| API docs | Versioned API contract |

## Architecture

Savio uses a modular monolith with explicit domain boundaries.

```text
React Web -> Go REST API -> Application Services -> Finance Engine -> PostgreSQL
                         -> Redis / Worker
                         -> MinIO
                         -> Bounded AI Provider
```

Financial records in PostgreSQL remain authoritative. Forecast and scenario calculations are deterministic. AI only interprets validated finance outputs and cannot silently mutate financial state.

Technical details:

- [System architecture](docs/architecture/system-architecture.md)
- [Frontend architecture](docs/architecture/frontend-architecture.md)
- [AI architecture](docs/architecture/ai-architecture.md)
- [Database design](docs/database/database-design.md)
- [Security model](docs/engineering/security.md)

## API

Application routes use `/api/v1`. Cookie authentication, CSRF validation, workspace authorization, consistent response envelopes, and centralized errors are enforced by the backend.

- [API contract](docs/api/api-contract.md)

## Take-Home Coverage

This repository implements the [full assignment specification](docs/assignment/take-home-test-specification.md).

| Requirement | Savio implementation | Detail |
| --- | --- | --- |
| Product and business flow | Connected personal-finance journey beyond CRUD | [Product foundation](docs/product/product-foundation.md) |
| Go REST API | Gin, GORM, PostgreSQL, `/api/v1` | [System architecture](docs/architecture/system-architecture.md) |
| React frontend | TypeScript, Tailwind CSS, Axios | [Frontend architecture](docs/architecture/frontend-architecture.md) |
| Authentication and security | HttpOnly cookies, CSRF, expiry, refresh rotation | [Security](docs/engineering/security.md) |
| Authorization | Backend-enforced `OWNER`, `MEMBER`, `VIEWER` roles | [Business requirements](docs/product/business-requirements.md) |
| Database and migrations | Explicit up/down migrations, constraints, indexes | [Database design](docs/database/database-design.md) |
| Validation and errors | Trust-boundary validation and centralized safe errors | [API contract](docs/api/api-contract.md) |
| Search and listing | Workspace-scoped search, filter, sort, pagination | [API contract](docs/api/api-contract.md) |
| UI/UX | Responsive loading, empty, error, and form states | [Frontend architecture](docs/architecture/frontend-architecture.md) |
| Testing | Backend, frontend, security, finance, and E2E strategy | [Testing strategy](docs/engineering/testing-strategy.md) |
| Bonus engineering | Redis worker, MinIO, Docker, AI, observability | [Implementation progress](docs/engineering/implementation-progress.md) |

## Technical Decisions

- **Modular monolith:** clear boundaries without premature distributed-system complexity.
- **Integer minor units:** authoritative money calculations avoid floating-point errors.
- **Derived balances:** account balances remain reconstructable from financial history.
- **Cookie authentication:** credentials never enter `localStorage` or `sessionStorage`.
- **Deterministic finance engine:** analytics, forecasts, and scenarios remain explainable and available without AI.
- **Bounded AI tools:** backend-injected authorization context prevents AI from choosing trusted scope.

Trade-offs: P0 uses one base currency per workspace, simple monthly category budgets, lightweight goals, and deterministic forecast assumptions. Advanced FX, investment tracking, autonomous AI actions, and production-scale observability remain outside current scope.

## Documentation

| Topic | Document |
| --- | --- |
| Product scope | [Product foundation](docs/product/product-foundation.md) |
| Business rules | [Business requirements](docs/product/business-requirements.md) |
| User journeys | [User flows](docs/product/user-flows.md) |
| Installation and operations | [Local development](docs/engineering/local-development.md) |
| System design | [System architecture](docs/architecture/system-architecture.md) |
| Frontend design | [Frontend architecture](docs/architecture/frontend-architecture.md) |
| AI design | [AI architecture](docs/architecture/ai-architecture.md) |
| Database | [Database design](docs/database/database-design.md) |
| API | [API contract](docs/api/api-contract.md) |
| Security | [Security](docs/engineering/security.md) |
| Testing | [Testing strategy](docs/engineering/testing-strategy.md) |
| Delivery status | [Implementation progress](docs/engineering/implementation-progress.md) |
| Roadmap | [Implementation plan](docs/engineering/implementation-plan.md) |
| Assignment | [Take-home specification](docs/assignment/take-home-test-specification.md) |

## Repository Layout

```text
backend/     Go API, worker, domain modules, migrations
frontend/    React application
docs/        Product, architecture, API, database, engineering docs
scripts/     Development and verification helpers
```

Current delivery status and future improvements are tracked in [implementation progress](docs/engineering/implementation-progress.md) and the [implementation plan](docs/engineering/implementation-plan.md).

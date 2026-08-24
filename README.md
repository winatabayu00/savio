# Savio

> **Understand Your Money. Plan What Comes Next.**

Savio is an **AI-powered personal cashflow intelligence and financial decision support platform** designed to help users understand their current financial condition, forecast what may happen next, and simulate the impact of financial decisions before making them.

Savio goes beyond traditional expense tracking.

Instead of only answering:

> "Where did my money go?"

Savio also helps answer:

> "Why did my cashflow change?"

> "What may happen next?"

> "What happens if I make this financial decision?"

The core product principle is:

> **Finance Engine calculates. AI interprets. User decides.**

---

# Table of Contents

- [1. Project Overview](#1-project-overview)
- [2. Problem Statement](#2-problem-statement)
- [3. Product Thesis](#3-product-thesis)
- [4. Target Users](#4-target-users)
- [5. Core Product Journey](#5-core-product-journey)
- [6. Core Features](#6-core-features)
- [7. Key Differentiators](#7-key-differentiators)
- [8. Technology Stack](#8-technology-stack)
- [9. System Architecture](#9-system-architecture)
- [10. Repository Structure](#10-repository-structure)
- [11. Financial Architecture](#11-financial-architecture)
- [12. Authentication & Security](#12-authentication--security)
- [13. Authorization & Workspace RBAC](#13-authorization--workspace-rbac)
- [14. AI Architecture](#14-ai-architecture)
- [15. Database Design Principles](#15-database-design-principles)
- [16. API Design](#16-api-design)
- [17. Prerequisites](#17-prerequisites)
- [18. Environment Configuration](#18-environment-configuration)
- [19. Local Development Setup](#19-local-development-setup)
- [20. Database Migrations](#20-database-migrations)
- [21. Seed Data](#21-seed-data)
- [22. Running the Backend](#22-running-the-backend)
- [23. Running the Frontend](#23-running-the-frontend)
- [24. Running the Worker](#24-running-the-worker)
- [25. Running with Docker Compose](#25-running-with-docker-compose)
- [26. Testing](#26-testing)
- [27. API Documentation](#27-api-documentation)
- [28. AI Configuration](#28-ai-configuration)
- [29. Graceful AI Degradation](#29-graceful-ai-degradation)
- [30. Observability](#30-observability)
- [31. Critical Business Invariants](#31-critical-business-invariants)
- [32. Technical Decisions](#32-technical-decisions)
- [33. Trade-Offs](#33-trade-offs)
- [34. MVP Scope](#34-mvp-scope)
- [35. Future Improvements](#35-future-improvements)
- [36. Demo Flow](#36-demo-flow)
- [37. Documentation](#37-documentation)
- [38. Engineering Principles](#38-engineering-principles)
- [39. Take-Home Assessment Coverage](#39-take-home-assessment-coverage)
- [40. Final Product Principle](#40-final-product-principle)

---

# 1. Project Overview

Savio is a personal finance application focused on:

```text
UNDERSTAND
    ↓
PREDICT
    ↓
SIMULATE
    ↓
EXPLAIN
    ↓
DECIDE
```

Traditional personal finance applications primarily help users record:

```text
income
expenses
accounts
budgets
```

Savio keeps those foundations but adds another layer:

```text
financial intelligence
```

The system uses deterministic financial calculations to understand current cashflow and project future conditions.

AI is then used to interpret those deterministic results in natural language.

---

# 2. Problem Statement

Many people already know approximately:

```text
how much they earn

how much they spend

how much money is left
```

but still struggle to answer:

```text
Why did my spending increase?

Which spending changes actually matter?

Will my cashflow remain healthy next month?

What happens if my income decreases?

What happens if I add a new installment?

Can I make a large purchase without destroying my financial buffer?

How does this decision affect my financial goal?
```

Traditional expense trackers are mostly retrospective.

They explain:

```text
what happened
```

but often do not help users understand:

```text
why it happened

what may happen next

what happens under a hypothetical decision
```

Savio is designed to close that gap.

---

# 3. Product Thesis

Savio's thesis is:

> Personal finance software becomes significantly more useful when it moves beyond recording history and helps users understand consequences before acting.

The product therefore combines:

```text
Financial Tracking
+
Deterministic Analytics
+
Cashflow Forecasting
+
Scenario Simulation
+
AI-Assisted Explanation
```

The responsibility boundaries are intentionally strict:

```text
Financial Records
      ↓
Finance Engine
      ↓
Forecast / Scenario Engine
      ↓
Structured Financial Facts
      ↓
AI Interpretation
      ↓
User Decision
```

---

# 4. Target Users

Savio primarily targets individuals who:

```text
receive regular or irregular income

manage several cash/bank/e-wallet accounts

want better visibility into cashflow

want to budget intentionally

are planning financial goals

frequently evaluate medium-to-large spending decisions
```

Representative users include:

```text
young professionals

freelancers

remote workers

students with income

small-business owners managing personal cashflow

people planning major purchases or lifestyle changes
```

---

# 5. Core Product Journey

The primary user journey is:

```text
TRACK
   ↓
UNDERSTAND
   ↓
FORECAST
   ↓
SIMULATE
   ↓
EXPLAIN
   ↓
DECIDE
   ↓
REVIEW
```

Example:

```text
User records salary and expenses
        ↓
Savio calculates monthly cashflow
        ↓
Savio detects higher dining spending
        ↓
Forecast projects the next 90 days
        ↓
User creates "Buy Laptop Rp15M" scenario
        ↓
Scenario Engine compares baseline vs scenario
        ↓
AI explains the main trade-offs
        ↓
User makes the final decision
```

---

# 6. Core Features

## Authentication

```text
Registration

Login

Logout

Logout All Sessions

Session Refresh

Current User

CSRF Protection
```

---

## Workspaces

Each user receives a default personal workspace.

P0 roles:

```text
OWNER

MEMBER

VIEWER
```

A user may later belong to more than one workspace.

---

## Accounts

Track:

```text
Cash

Bank

E-Wallet

Savings
```

Each account belongs to a workspace.

Account balance is derived from financial history rather than treated as an independently editable source of truth.

---

## Categories

Categories support:

```text
INCOME

EXPENSE
```

Savio provides system categories and may also support workspace-specific custom categories.

---

## Transactions

Supported transaction types:

```text
INCOME

EXPENSE

ADJUSTMENT
```

Financial history remains auditable.

Posted financial effects must not disappear silently.

---

## Transfers

Move money between accounts atomically.

Transfers:

```text
do not count as income

do not count as expense

do not change total internal portfolio value
```

---

## Account Reconciliation

If tracked balance differs from real balance:

```text
Tracked Balance
Rp4.800.000

Actual Balance
Rp5.000.000

Difference
+Rp200.000
```

Savio records an adjustment rather than overwriting financial history.

---

## Recurring Transactions

Model expected:

```text
salary

rent

subscriptions

installments

regular bills
```

Recurring rules create planned occurrences.

They do not automatically become actual transactions by default.

The user confirms the occurrence unless auto-posting is explicitly enabled.

---

## Budgets

Create category budgets such as:

```text
Food & Dining
Rp2.000.000 / month
```

Savio calculates:

```text
actual spending

remaining amount

utilization

projected month-end spend

budget risk
```

---

## Goals

Basic financial goals support:

```text
target amount

target date

tracked progress

required contribution

simple feasibility
```

---

## Analytics

Analytics include:

```text
Income

Expense

Net Cashflow

Savings Rate

Category Breakdown

Period Comparison

Spending Changes

Recurring Expense Summary
```

---

## Cashflow Forecast

Savio projects future cashflow using deterministic inputs such as:

```text
current financial state

recurring income

recurring expenses

historical variable spending

explicit assumptions
```

Forecast outputs may include:

```text
projected ending balance

minimum projected balance

minimum-balance date

projected income

projected expenses

event timeline

confidence

assumptions
```

---

## Scenario Simulator

Users can simulate decisions such as:

```text
Buy a Rp15M laptop

Income decreases by 30%

Resign next month

Add a monthly installment

Increase monthly savings

Reduce recurring spending
```

A scenario is a:

> **Non-destructive overlay over the baseline forecast.**

It never changes actual financial records.

---

## AI Transaction Categorization

Savio may suggest a category based on:

```text
merchant

transaction description

available categories
```

The user remains free to accept or reject the suggestion.

---

## AI Insights

Deterministic financial signals may trigger AI explanations.

Example:

```text
Food & Dining spending:
+60% vs recent baseline
```

The deterministic engine identifies the change.

AI explains the meaning.

---

## Savio Copilot

Copilot provides grounded natural-language access to financial intelligence.

Example questions:

```text
Why did I spend more this month?

Where did my money go?

What are my largest recurring expenses?

Which budget is most at risk?

What does my forecast look like?

What happens if I buy a Rp15M laptop?

What happens if my income drops by 30%?
```

Copilot works through bounded finance tools rather than unrestricted database access.

---

# 7. Key Differentiators

Savio is intentionally differentiated through three major capabilities.

## 7.1 Explainable AI Insights

AI does not independently invent financial conclusions.

Architecture:

```text
Deterministic Signal
      ↓
Structured Financial Facts
      ↓
AI Explanation
```

---

## 7.2 Cashflow Forecasting

Savio projects future financial conditions from known and estimated cashflow.

The forecast is:

```text
deterministic

assumption-driven

explainable
```

not an LLM prediction.

---

## 7.3 Financial Decision Simulator

Users can test financial decisions before making them.

Example:

```text
Baseline
vs
Buy Laptop Scenario
```

and compare:

```text
Ending Balance

Minimum Balance

Cash Runway

Savings Rate

Goal Impact
```

---

# 8. Technology Stack

## Backend

```text
Go

Gin

GORM

PostgreSQL

Explicit SQL Migrations
```

---

## Frontend

```text
React

TypeScript

Vite

Tailwind CSS

React Router

Axios

TanStack Query

React Hook Form

Zod
```

---

## Infrastructure

```text
PostgreSQL

Redis

MinIO

Docker Compose
```

---

## Background Processing

```text
Go Worker

Redis-backed Queue
```

A library such as Asynq may be used.

---

## AI

```text
AI Provider Abstraction

OpenAI-Compatible Provider

Mock Provider

Structured Output Validation

Domain-Specific AI Tools
```

---

## Testing

Backend:

```text
Go testing

Testify
```

Frontend:

```text
Vitest

React Testing Library

MSW

Playwright
```

---

# 9. System Architecture

High-level architecture:

```text
                         USER
                          │
                          ▼
                 ┌─────────────────┐
                 │ React Frontend  │
                 └────────┬────────┘
                          │
                          │ HTTPS / REST
                          ▼
                 ┌─────────────────┐
                 │     Go API      │
                 │   Gin + GORM    │
                 └────────┬────────┘
                          │
            ┌─────────────┼──────────────┐
            │             │              │
            ▼             ▼              ▼
      Domain Services  Finance Engine  AI Orchestrator
            │             │              │
            └──────┬──────┘              │
                   │                     ▼
                   ▼               AI Provider
              PostgreSQL
                   │
                   │ async jobs
                   ▼
                 Redis
                   │
                   ▼
               Go Worker
                   │
            ┌──────┴──────┐
            ▼             ▼
       PostgreSQL        MinIO
```

Savio uses a:

```text
MODULAR MONOLITH
```

rather than microservices.

This keeps:

```text
transaction boundaries simple

deployment understandable

local development reproducible

business rules centralized
```

while maintaining clear internal module boundaries.

---

# 10. Repository Structure

Target repository:

```text
savio/
├── README.md
├── AGENTS.md
├── DESIGN.md
├── Makefile
├── docker-compose.yml
├── .env.example
│
├── backend/
│   ├── cmd/
│   │   ├── api/
│   │   └── worker/
│   │
│   ├── internal/
│   │   ├── platform/
│   │   ├── auth/
│   │   ├── users/
│   │   ├── workspaces/
│   │   ├── accounts/
│   │   ├── categories/
│   │   ├── transactions/
│   │   ├── transfers/
│   │   ├── recurring/
│   │   ├── budgets/
│   │   ├── goals/
│   │   ├── analytics/
│   │   ├── forecast/
│   │   ├── scenarios/
│   │   ├── ai/
│   │   ├── notifications/
│   │   └── audit/
│   │
│   ├── migrations/
│   ├── seeds/
│   └── tests/
│
├── frontend/
│   ├── src/
│   │   ├── app/
│   │   ├── features/
│   │   └── shared/
│   └── tests/
│
├── docs/
│   ├── product/
│   ├── database/
│   ├── api/
│   ├── architecture/
│   ├── engineering/
│   └── assignment/
│
├── docker/
├── scripts/
└── .github/
    └── workflows/
```

---

# 11. Financial Architecture

Savio separates:

```text
financial records
```

from:

```text
derived financial intelligence
```

Authority hierarchy:

```text
Authoritative Ledger
        ↓
Finance Engine
        ↓
Analytics
        ↓
Forecast
        ↓
Scenario
        ↓
AI Interpretation
```

---

## Money Representation

Authoritative money should use:

```text
integer minor units
```

stored with:

```text
BIGINT
```

plus ISO currency.

Conceptually:

```text
amount_minor = 1500000
currency = IDR
```

The application must never use:

```text
float32

float64
```

for authoritative financial arithmetic.

For currencies with fractional minor units, the same architecture remains valid through the currency's defined minor-unit scale.

---

## Workspace Currency

P0 uses:

```text
one base currency per workspace
```

Accounts in the workspace must use the same supported currency.

Multi-currency conversion is intentionally deferred.

---

## Account Balance

Account balance is derived from:

```text
opening_balance_minor
+
posted financial ledger effects
```

It must not be treated as an independently mutable field.

If a cached/materialized balance is introduced later, it is:

```text
derived

reconcilable

non-authoritative by itself
```

---

# 12. Authentication & Security

Authentication uses:

```text
secure cookies
```

rather than browser token storage.

Savio does not store authentication credentials in:

```text
localStorage

sessionStorage
```

---

## Authentication Model

```text
Short-Lived Access Token
+
Rotating Refresh Token
+
Server-Side Session
```

Recommended:

```text
Access:
JWT

Refresh:
Opaque Cryptographically Random Token
```

---

## Cookies

Production cookies use:

```text
HttpOnly

Secure

SameSite

Path

Max-Age / Expires
```

---

## Refresh Rotation

Each successful refresh:

```text
validates current refresh token
↓
rotates token
↓
stores new token hash
↓
invalidates previous token
```

Concurrent rotation is protected server-side.

---

## CSRF

Because authentication uses cookies, CSRF protection is mandatory.

State-changing methods:

```text
POST

PUT

PATCH

DELETE
```

must carry a valid CSRF token according to the configured signed double-submit/session-bound strategy.

---

## Passwords

Passwords are stored using a strong password hashing algorithm such as:

```text
Argon2id
```

Raw passwords are never persisted.

---

## Rate Limiting

Rate limiting protects:

```text
Login

Registration

Refresh

AI Categorization

Savio Copilot
```

Redis may be used for distributed counters.

---

# 13. Authorization & Workspace RBAC

Savio uses workspace-scoped authorization.

P0 roles:

```text
OWNER

MEMBER

VIEWER
```

---

## Permission Overview

| Capability | OWNER | MEMBER | VIEWER |
| --- | --- | --- | --- |
| View Financial Data | Yes | Yes | Yes |
| Create Transactions | Yes | Yes | No |
| Update Finance Records | Yes | Yes | No |
| Create Budgets | Yes | Yes | No |
| Create Scenarios | Yes | Yes | No |
| View Forecast | Yes | Yes | Yes |
| Use Copilot | Yes | Yes | Yes |
| Manage Members | Yes | No | No |
| Change Roles | Yes | No | No |
| Manage Workspace | Yes | No | No |

Authorization is enforced on the backend.

Hiding a button in the frontend is not considered authorization.

---

## Workspace Scope

Resource lookup must include:

```text
workspace_id
```

Example:

```text
transaction.id = requested_id
AND
transaction.workspace_id = authorized_workspace
```

This prevents IDOR and cross-workspace access.

---

## Last Owner Protection

Savio must prevent:

```text
removing the last OWNER

demoting the last OWNER
```

without transferring ownership.

---

# 14. AI Architecture

The AI architecture follows:

> **Finance Engine calculates. AI interprets. User decides.**

AI is not part of the authoritative financial path.

---

## Allowed AI Responsibilities

```text
Transaction categorization suggestion

Financial pattern explanation

Forecast explanation

Scenario explanation

Natural-language financial Q&A
```

---

## Forbidden AI Responsibilities

AI must not independently determine authoritative:

```text
Account Balance

Income

Expense

Savings Rate

Budget Utilization

Forecast

Scenario Calculation

Goal Progress
```

---

## AI Copilot Architecture

```text
User Question
      ↓
AI Orchestrator
      ↓
Intent / Tool Selection
      ↓
Bounded Finance Tools
      ↓
Finance Services
      ↓
Structured Facts
      ↓
Minimal Context Builder
      ↓
LLM
      ↓
Structured Output Validation
      ↓
Response
```

---

## AI Tools

Examples:

```text
get_cashflow_summary

compare_periods

get_category_breakdown

get_spending_changes

get_recurring_expenses

get_budget_status

get_goal_status

get_forecast

calculate_scenario
```

The model has no unrestricted:

```text
SQL

shell

filesystem

generic network
```

access.

---

## AI Security

The model never chooses trusted:

```text
user_id

workspace_id

role
```

Authorization context is injected by the backend.

Model output is always treated as untrusted until validated.

---

# 15. Database Design Principles

PostgreSQL is the authoritative database.

Database principles:

```text
explicit migrations

foreign keys

unique constraints

check constraints

intentional indexes

workspace scoping

financial integrity

reproducible empty-database setup
```

---

## Migrations

Production schema changes use:

```text
explicit SQL migrations
```

rather than GORM `AutoMigrate` as the source of truth.

Each migration should provide:

```text
up

down
```

where safely reversible.

---

## Core Entities

Expected entities include:

```text
users

user_settings

auth_sessions

workspaces

workspace_memberships

accounts

categories

transactions

transfers

recurring_rules

recurring_occurrences

budgets

goals

forecast_snapshots

financial_scenarios

scenario_modifications

scenario_snapshots

ai_insights

ai_requests

audit_logs
```

---

# 16. API Design

All application endpoints use:

```text
/api/v1
```

Infrastructure endpoints may remain:

```text
/health

/ready
```

---

## Success Response

Typical response:

```json
{
  "success": true,
  "data": {}
}
```

---

## Error Response

Typical:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "details": {}
  },
  "message": "Please check the submitted data."
}
```

---

## HTTP Status Codes

Savio uses appropriate HTTP semantics:

```text
200 OK

201 Created

204 No Content

400 Bad Request

401 Unauthorized

403 Forbidden

404 Not Found

409 Conflict

422 Unprocessable Entity

429 Too Many Requests

500 Internal Server Error

503 Service Unavailable
```

---

## Collection Endpoints

Lists should support where relevant:

```text
search

filter

sort

pagination
```

Pagination uses bounded limits.

Example:

```text
page = 1

limit = 20

maximum limit = 100
```

---

# 17. Prerequisites

Recommended local prerequisites:

```text
Docker

Docker Compose

Go

Node.js

npm / pnpm

Make
```

Example development versions:

```text
Go 1.24+

Node.js 22+

PostgreSQL 16+

Redis 7+
```

Use the versions defined in the repository configuration as the final source of truth.

---

# 18. Environment Configuration

Copy:

```bash
cp .env.example .env
```

Example configuration:

```env
APP_ENV=development
APP_PORT=8080

DATABASE_URL=postgres://savio:savio@localhost:5432/savio?sslmode=disable

REDIS_URL=redis://localhost:6379

JWT_SECRET=replace-with-development-secret
CSRF_SECRET=replace-with-development-secret

ACCESS_TOKEN_TTL_MINUTES=15
REFRESH_TOKEN_TTL_DAYS=7

FRONTEND_ORIGIN=http://localhost:5173

AI_ENABLED=false
AI_PROVIDER=mock
AI_BASE_URL=
AI_API_KEY=
AI_MODEL=
AI_TIMEOUT_SECONDS=20

MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minio
MINIO_SECRET_KEY=minio-development-password
MINIO_BUCKET=savio
MINIO_USE_SSL=false
```

Never commit real production credentials.

---

# 19. Local Development Setup

**Fastest start — one command** (infra + DB + demo data + backend + frontend):

```bash
git clone <repository-url>
cd savio
./scripts/dev.sh
# open http://localhost:5173  (demo: demo@savio.test / DemoPassword!23)
```

The script creates `.env` from `.env.example`, boots PostgreSQL/Redis/MinIO via
Docker, runs migrations, seeds demo finance data, and starts the API (8080) and
the web app (5173). Stop everything with `Ctrl-C`.

Manual breakdown (equivalent steps):

Clone the repository:

```bash
git clone <repository-url>
cd savio
```

Start infrastructure:

```bash
make infra-up
```

or:

```bash
docker compose up -d postgres redis minio
```

Install backend dependencies:

```bash
cd backend
go mod download
```

Install frontend dependencies:

```bash
cd ../frontend
npm install
```

Return to project root:

```bash
cd ..
```

Run migrations:

```bash
make migrate-up
```

Optionally seed development data:

```bash
make seed-demo
```

---

# 20. Database Migrations

Run all migrations:

```bash
make migrate-up
```

Rollback one migration:

```bash
make migrate-down
```

Exact underlying migration commands may be:

```bash
migrate -path backend/migrations \
  -database "$DATABASE_URL" up
```

Rollback:

```bash
migrate -path backend/migrations \
  -database "$DATABASE_URL" down 1
```

---

## Fresh Database Verification

A fresh environment must support:

```text
empty PostgreSQL
↓
run migrations
↓
run seed if desired
↓
start application
```

No manual schema modification should be required.

---

# 21. Seed Data

Development seed may provide:

```text
system categories

demo user

demo workspace

accounts

salary

expenses

recurring bills

budget

goal

forecast-friendly history
```

Run:

```bash
make seed-demo
```

Demo financial information must be synthetic.

Automated tests do not depend on demo seed data.

---

# 22. Running the Backend

From repository root:

```bash
make dev-api
```

or:

```bash
cd backend
go run ./cmd/api
```

Default API:

```text
http://localhost:8080
```

Health:

```text
GET /health
```

Readiness:

```text
GET /ready
```

---

# 23. Running the Frontend

From repository root:

```bash
make dev-web
```

or:

```bash
cd frontend
npm run dev
```

Default frontend:

```text
http://localhost:5173
```

---

# 24. Running the Worker

From repository root:

```bash
make dev-worker
```

or:

```bash
cd backend
go run ./cmd/worker
```

The worker may process:

```text
recurring occurrence generation

background financial signals

AI insight generation

notifications

session cleanup
```

depending on implemented scope.

---

# 25. Running with Docker Compose

To run the full local environment:

```bash
docker compose up --build
```

Expected services:

```text
frontend

api

worker

postgres

redis

minio
```

Stop:

```bash
docker compose down
```

Remove development volumes:

```bash
docker compose down -v
```

Use volume removal carefully because it deletes local database/object-storage data.

---

# 26. Testing

Run all tests:

```bash
make test
```

---

## Backend Tests

```bash
make test-backend
```

or:

```bash
cd backend
go test ./...
```

---

## Race Detector

```bash
cd backend
go test -race ./...
```

---

## Frontend Tests

```bash
make test-frontend
```

or:

```bash
cd frontend
npm run test
```

---

## Type Check

```bash
cd frontend
npm run typecheck
```

---

## Frontend Build

```bash
cd frontend
npm run build
```

---

## End-to-End Tests

```bash
make test-e2e
```

or:

```bash
cd frontend
npm run test:e2e
```

---

## Critical Automated Coverage

The project prioritizes tests for:

```text
financial calculations

workspace authorization

IDOR protection

CSRF

refresh token rotation

concurrent refresh

transaction integrity

transfer integrity

recurring idempotency

forecast calculations

scenario isolation

AI tool authorization

AI output validation

Axios single-flight refresh
```

---

# 27. API Documentation

Primary API documentation:

```text
docs/api/api-contract.md
```

Machine-readable OpenAPI:

```text
docs/api/openapi.yaml
```

If Swagger UI is enabled:

```text
/api/docs
```

The implementation, API contract, OpenAPI schema, and frontend types should remain synchronized.

---

# 28. AI Configuration

Savio supports an AI provider abstraction.

Example real-provider configuration:

```env
AI_ENABLED=true

AI_PROVIDER=openai_compatible

AI_BASE_URL=https://provider.example.com/v1

AI_API_KEY=...

AI_MODEL=...

AI_TIMEOUT_SECONDS=20
```

---

## Mock AI Provider

For development or automated testing:

```env
AI_ENABLED=true
AI_PROVIDER=mock
```

This allows AI workflows to be exercised without external provider credentials.

CI should normally use:

```text
mock provider
```

rather than a live AI API.

---

# 29. Graceful AI Degradation

Savio's deterministic financial features do not depend on AI availability.

If the provider is unavailable:

```text
Accounts              WORK

Transactions          WORK

Transfers             WORK

Recurring             WORK

Budgets               WORK

Analytics             WORK

Forecast              WORK

Scenario Simulator    WORK

AI Categorization     DEGRADED

AI Insights           DEGRADED

Savio Copilot         DEGRADED
```

This separation is intentional.

---

# 30. Observability

Savio includes lightweight production-minded observability.

Minimum:

```text
Structured Logs

Request ID

Health Endpoint

Readiness Endpoint

Worker Logs

AI Request Metadata
```

---

## Structured Log Fields

Examples:

```text
timestamp

level

request_id

user_id

workspace_id

method

path

status

duration_ms

error_code
```

Sensitive financial request bodies should not be logged by default.

---

## AI Operational Metadata

May include:

```text
provider

model

feature

latency

input token count

output token count

status

error code
```

Raw sensitive prompts should not be logged by default.

---

# 31. Critical Business Invariants

The implementation must preserve these invariants.

```text
INV-001
A user may access only workspaces where they have active membership.

INV-002
VIEWER cannot mutate workspace financial state.

INV-003
The final OWNER cannot be removed or demoted without ownership transfer.

INV-004
Money uses exact integer minor-unit arithmetic.

INV-005
Account balance is reconstructable from opening balance and posted financial effects.

INV-006
Transfers do not change total internal portfolio value.

INV-007
Transfers do not count as income or expense.

INV-008
Voided financial records do not count as active income or expense.

INV-009
Reconciliation creates an adjustment instead of rewriting account history.

INV-010
A recurring occurrence becomes actual at most once.

INV-011
Forecast is deterministic from its input state and documented assumptions.

INV-012
Scenario calculation never mutates real financial state.

INV-013
AI never selects trusted authorization identity.

INV-014
AI does not become the source of authoritative financial calculations.

INV-015
AI-originated mutations require validation and explicit user confirmation.

INV-016
Authentication credentials are not stored in localStorage or sessionStorage.

INV-017
Cookie-authenticated state-changing operations are protected by CSRF controls.

INV-018
Refresh rotation invalidates the previous refresh token.

INV-019
Cross-workspace IDs cannot expose private financial data.

INV-020
Backend validation remains the source of truth.
```

---

# 32. Technical Decisions

## Modular Monolith

Decision:

```text
Modular Monolith
```

Reason:

```text
clear domain boundaries

simple transactional consistency

simple local development

low deployment complexity

good take-home reviewability
```

---

## Explicit Database Migrations

Decision:

```text
SQL migrations
```

instead of relying on runtime GORM AutoMigrate.

Reason:

```text
reproducibility

reviewability

rollback

schema history
```

---

## Integer Minor Units for Money

Decision:

```text
BIGINT minor units
+
ISO currency
```

Reason:

```text
exact arithmetic

no floating-point precision errors

simple deterministic calculations
```

---

## Derived Account Balance

Decision:

```text
ledger-derived balance
```

instead of mutable balance as independent source of truth.

Reason:

```text
auditability

reconstructability

safer corrections

reduced drift risk
```

---

## Secure Cookie Authentication

Decision:

```text
HttpOnly cookie authentication
```

instead of browser bearer-token storage.

Reason:

```text
reduces token exposure to JavaScript

satisfies assignment requirement
```

Cost:

```text
requires deliberate CSRF protection
```

---

## Rotating Opaque Refresh Tokens

Decision:

```text
opaque server-tracked refresh sessions
```

Reason:

```text
revocation

rotation

session management

logout-all support
```

---

## Workspace RBAC

Decision:

```text
OWNER / MEMBER / VIEWER
```

Reason:

```text
meaningful real authorization levels

stronger than cosmetic frontend roles

future-compatible with shared finance
```

---

## Deterministic Forecast

Decision:

```text
rule-based deterministic forecast
```

instead of AI-generated financial predictions.

Reason:

```text
testability

explainability

reproducibility
```

---

## Non-Destructive Scenario Engine

Decision:

```text
scenario overlay
```

rather than modifying real finance state.

Reason:

```text
safety

repeatability

clear what-if analysis
```

---

## Bounded AI Tools

Decision:

```text
domain-specific AI tools
```

instead of unrestricted SQL or autonomous agents.

Reason:

```text
security

authorization

grounding

testability
```

---

# 33. Trade-Offs

## No Microservices

Savio deliberately avoids microservices in the initial implementation.

The product does not yet require independently deployed distributed services.

---

## No Full Accounting Double-Entry System

Savio is a personal cashflow intelligence product rather than accounting software.

The ledger model preserves required financial integrity without implementing a complete general-ledger accounting product.

---

## Single Workspace Currency

P0 uses one base currency.

Multi-currency would require:

```text
FX rates

conversion timestamps

rate sources

conversion rules
```

and is deferred.

---

## Recurring Confirmation by Default

A scheduled recurring item does not automatically mean money really moved.

Therefore:

```text
planned occurrence
```

and:

```text
actual posted transaction
```

remain separate.

---

## AI Is Optional

The product remains useful without AI.

This increases architectural separation but significantly improves reliability.

---

## No Vector Database

Core Savio AI workflows use structured financial tools.

A vector database is unnecessary for P0.

---

## No Generic Autonomous Agent

Savio prioritizes:

```text
bounded orchestration
```

over broad autonomous execution.

This provides better safety and predictability.

---

# 34. MVP Scope

P0 target:

```text
Authentication

Cookie Sessions

Refresh Rotation

CSRF

Workspace RBAC

Accounts

Categories

Transactions

Transfers

Reconciliation

Recurring Planning

Budgets

Basic Goals

Dashboard

Analytics

Search

Filter

Sort

Pagination

Cashflow Forecast

Scenario Simulator

AI Categorization

AI Insights

Grounded Copilot

Scenario Explanation

Audit Logging

Rate Limiting

Docker Compose

Tests

API Documentation
```

---

# 35. Future Improvements

Potential P1/P2 capabilities:

```text
Notification Center

Advanced Goal Planning

CSV Import

Receipt Upload

OCR

Report Export

Advanced Analytics

Household Collaboration

Shared Goals

Shared Budgets

Persistent Copilot Conversations

Semantic Transaction Search

Bank Integration

Multi-Currency

Advanced Forecast Models

Provider Fallback

AI Evaluation Pipeline

OpenTelemetry

Prometheus Metrics
```

These are intentionally secondary to P0 correctness.

---

# 36. Demo Flow

Recommended project demonstration:

```text
1. Register a new account.

2. Savio creates a personal workspace and OWNER membership.

3. Add a bank account.

4. Add salary income.

5. Add recurring rent.

6. Add several expenses.

7. Savio suggests a category for a transaction.

8. Create a Food & Dining budget.

9. Open Dashboard.

10. Review:
    - total balance
    - income
    - expense
    - net cashflow
    - category behavior

11. Open Cashflow Forecast.

12. Review:
    - projected ending balance
    - minimum balance
    - recurring timeline
    - assumptions

13. Create scenario:
    "Buy Laptop"

14. Add:
    One-time expense Rp15M.

15. Calculate.

16. Compare:
    Baseline vs Scenario.

17. Review:
    - balance impact
    - runway impact
    - goal impact

18. Ask Savio AI to explain the scenario.

19. Open Copilot.

20. Ask:
    "Why did I spend more this month?"

21. Copilot retrieves deterministic financial facts.

22. AI explains the primary spending drivers.

23. The user makes the final decision.
```

---

## RBAC Demo

```text
1. OWNER adds another existing user as VIEWER.

2. VIEWER opens Dashboard successfully.

3. VIEWER attempts POST /transactions.

4. Backend returns 403.

5. OWNER changes VIEWER to MEMBER.

6. MEMBER can now create a transaction.
```

This demonstrates genuine backend authorization.

---

## AI Failure Demo

Set AI provider unavailable.

Expected:

```text
Dashboard works.

Transactions work.

Budgets work.

Forecast works.

Scenario calculation works.

Copilot displays graceful degradation.
```

---

# 37. Documentation

All source-of-truth documentation lives in the repository so agents and humans share one reference.

| Document | Purpose |
| --- | --- |
| [README.md](README.md) | This file — project overview and setup. |
| [AGENTS.md](AGENTS.md) | Coding-agent execution rules and engineering guardrails. |
| [DESIGN.md](DESIGN.md) | Visual design system and UX guidelines. |
| [docs/product/product-foundation.md](docs/product/product-foundation.md) | Product vision, positioning, target users, scope, and principles. |
| [docs/product/business-requirements.md](docs/product/business-requirements.md) | Detailed business rules and requirements (authoritative). |
| [docs/product/user-flows.md](docs/product/user-flows.md) | End-to-end user workflows and state transitions. |
| [docs/database/database-design.md](docs/database/database-design.md) | Database architecture, entities, relationships, constraints, and indexes. |
| [docs/api/api-contract.md](docs/api/api-contract.md) | REST API contract and response conventions. |
| [docs/architecture/system-architecture.md](docs/architecture/system-architecture.md) | Application architecture and infrastructure boundaries. |
| [docs/architecture/frontend-architecture.md](docs/architecture/frontend-architecture.md) | Frontend routes, feature organization, state management, and UI architecture. |
| [docs/architecture/ai-architecture.md](docs/architecture/ai-architecture.md) | AI responsibilities, provider abstraction, orchestration, tools, and guardrails. |
| [docs/engineering/security.md](docs/engineering/security.md) | Authentication, CSRF, RBAC, privacy, AI security, and hardening requirements. |
| [docs/engineering/testing-strategy.md](docs/engineering/testing-strategy.md) | Unit, integration, security, concurrency, frontend, AI, and E2E testing strategy. |
| [docs/engineering/implementation-plan.md](docs/engineering/implementation-plan.md) | Milestone-based implementation roadmap. |
| [docs/assignment/take-home-test-specification.md](docs/assignment/take-home-test-specification.md) | Normalized copy/reference of the original take-home requirements. |

Documentation precedence and conflict-resolution rules are defined in [AGENTS.md](AGENTS.md#5-documentation-precedence).

---

# 38. Engineering Principles

Savio follows several engineering principles.

---

## Correctness Before Complexity

Prefer:

```text
simple deterministic implementation
```

over unnecessary infrastructure sophistication.

---

## Explicit Over Clever

Prefer:

```text
TransferService.Create()
```

over generic abstractions hiding important financial rules.

---

## Backend as Source of Truth

Frontend improves experience.

It does not determine authoritative financial state.

---

## Financial Integrity Before Convenience

If a shortcut can corrupt or obscure financial history:

```text
do not take the shortcut.
```

---

## Security Is a Backend Responsibility

Frontend permissions improve UX.

Backend enforcement provides security.

---

## Derived Data Must Be Reproducible

Examples:

```text
balance

analytics

forecast

scenario output
```

must be derivable from authoritative state and explicit assumptions.

---

## AI Is Bounded

AI operates only within explicit application capabilities.

It cannot create new privileges.

---

## Graceful Degradation

Optional infrastructure failure should degrade only the related capability.

---

## Tests Focus on Risk

Highest-value tests protect:

```text
financial integrity

security

authorization

concurrency

forecast correctness

scenario isolation

AI boundaries
```

---

# 39. Take-Home Assessment Coverage

Savio is intentionally designed to demonstrate the required fullstack engineering capabilities.

## Backend Architecture

Demonstrated through:

```text
Go

Gin

GORM

Modular Monolith

Handler → Service → Repository

Deterministic Finance Engine
```

---

## API & Business Logic

Demonstrated through:

```text
accounts

transactions

transfers

recurring lifecycle

budgets

forecast

scenario simulation

business-rule validation
```

---

## Authentication & Security

Demonstrated through:

```text
cookie auth

CSRF

refresh rotation

server-side sessions

RBAC

workspace authorization

rate limiting

secure headers
```

---

## UI / UX

Demonstrated through:

```text
custom application shell

responsive dashboard

transaction workflows

forecast visualization

scenario comparison

AI insight presentation

loading/empty/error states
```

---

## Database & Migrations

Demonstrated through:

```text
PostgreSQL

explicit migrations

rollback

foreign keys

unique constraints

check constraints

indexes
```

---

## Frontend Architecture

Demonstrated through:

```text
React

TypeScript

feature modules

TanStack Query

React Hook Form

Zod

Axios interceptors
```

---

## Testing

Demonstrated through:

```text
finance unit tests

PostgreSQL integration tests

security tests

concurrency tests

AI mock tests

frontend integration tests

E2E tests
```

---

## Documentation

Demonstrated through:

```text
product documentation

architecture documentation

database documentation

API documentation

security documentation

testing documentation

implementation roadmap
```

---

## Engineering Bonus

Potentially demonstrated through:

```text
Redis queue

MinIO

Docker Compose

AI integration

rate limiting

optimistic locking

request ID

structured logs

health/readiness

CI
```

---

# 40. Final Product Principle

Savio is not designed to replace user judgment.

It is designed to provide better information before judgment is required.

The hierarchy is:

```text
SECURITY
    ↓
AUTHORIZATION
    ↓
AUTHORITATIVE FINANCIAL DATA
    ↓
DETERMINISTIC FINANCE ENGINE
    ↓
FORECAST
    ↓
SCENARIO
    ↓
AI INTERPRETATION
    ↓
USER DECISION
```

The product can therefore be summarized in one sentence:

> **Savio helps users understand their financial position today, see what may happen next, and test financial decisions before making them.**

And the engineering rule remains:

> **Finance Engine calculates. AI interprets. User decides.**
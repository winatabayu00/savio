# Savio — Implementation Plan

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Business Requirements](../product/business-requirements.md) — the rules each milestone implements.
- [Testing Strategy](testing-strategy.md) — test gates per milestone.
- [System Architecture](../architecture/system-architecture.md) — module layout this plan sequences.
- [Take-Home Test Specification](../assignment/take-home-test-specification.md) — acceptance constraints.

## 1. Document Purpose

This document defines the implementation plan for Savio.

The purpose of this document is to convert Savio's completed product and engineering documentation into an executable development roadmap.

The implementation plan translates:

```text
Product Foundation
↓
Business Requirements
↓
User Flows
↓
Database Design
↓
API Contract
↓
AI Architecture
↓
System Architecture
↓
Frontend Architecture
↓
Design System
↓
Security Architecture
↓
Testing Strategy
```

into:

```text
IMPLEMENTABLE MILESTONES
```

The objective is not simply to finish as many features as possible.

The objective is to deliver a coherent, production-minded take-home project with strong implementation quality in the areas that matter most:

```text
Backend Architecture

Business Logic

Authentication & Security

UI / UX

Database & Migrations

Frontend Architecture

Testing

Documentation

Git Quality
```

The central Savio product rule remains:

> **Finance Engine calculates. AI interprets. User decides.**

---

# 2. Implementation Strategy

Savio should be implemented:

```text
foundation first
↓
financial correctness
↓
planning intelligence
↓
AI
↓
polish
```

Do not begin implementation from:

```text
AI Copilot
```

or:

```text
dashboard visual polish
```

before the financial domain is stable.

---

# 3. Primary Engineering Principle

Implementation priority must follow dependency direction.

```text
INFRASTRUCTURE
       ↓
SECURITY
       ↓
AUTHORITATIVE FINANCIAL DOMAIN
       ↓
DETERMINISTIC FINANCE ENGINE
       ↓
FORECAST
       ↓
SCENARIO
       ↓
AI
       ↓
FRONTEND POLISH
```

---

# 4. Implementation Philosophy

Every milestone should produce:

```text
working code
+
tests
+
documentation alignment
```

Avoid:

```text
build all backend first
then test everything later
```

Preferred:

```text
implement module
↓
test module
↓
integrate module
↓
move forward
```

---

# 5. Scope Discipline

Savio has many potential features.

The implementation should protect the core product thesis first:

```text
Track
↓
Understand
↓
Forecast
↓
Simulate
↓
Explain
↓
Decide
```

Features unrelated to that chain should not delay submission.

---

# 6. P0 Product Scope

P0 represents the required implementation target.

```text
Authentication

CSRF

Session Refresh

Authorization / RBAC

Accounts

Categories

Transactions

Transfers

Recurring Transactions

Budgets

Basic Goals

Dashboard

Analytics

Forecast

Scenario Simulator

AI Categorization

AI Insights

AI Copilot

Scenario AI Explanation

Search

Filter

Sort

Pagination

Audit

Health

Docker

Tests

Documentation
```

---

# 7. P1 Scope

Implement only after P0 is stable.

```text
Session Management UI

Notifications

AI Insight Feedback

Background AI Insight Generation

Advanced Goal Planning

Advanced Reports

Additional Analytics

AI Conversation Persistence
```

---

# 8. P2 Scope

Do not implement unless core project is complete.

```text
Receipt Upload

OCR

CSV Import

Exports

Household Collaboration

Advanced Workspace UX

Multi-Currency

Bank Integration

Dark Mode

Advanced AI Memory

Semantic Search

Model Routing
```

---

# 9. Important Scope Correction — Authorization

Although Savio is primarily a personal finance product, the take-home requires meaningful multi-level authorization.

Therefore P0 must include a minimal authorization model that is stronger than:

```text
USER
ADMIN
```

alone.

Recommended P0 model:

```text
Personal Workspace
+
Workspace Membership
+
Workspace Roles
```

Roles:

```text
OWNER

MEMBER

VIEWER
```

Optional system role:

```text
PLATFORM_ADMIN
```

This allows the project to demonstrate meaningful backend-enforced authorization without turning Savio into a collaboration-heavy product.

---

# 10. P0 Workspace Model

Each new user receives:

```text
one default personal workspace
```

during registration.

Example:

```text
Alex's Workspace
```

Registration transaction:

```text
Create User
↓
Create Workspace
↓
Create OWNER Membership
↓
Create User Settings
↓
Create Default Categories if required
↓
Create Auth Session
```

All must succeed or roll back together where practical.

---

# 11. Workspace Roles

P0 roles:

```text
OWNER
MEMBER
VIEWER
```

OWNER:

```text
full workspace finance access

manage workspace

manage members

change member roles
```

MEMBER:

```text
create/update financial resources

view financial intelligence

cannot manage ownership
```

VIEWER:

```text
read-only financial access
```

---

# 12. Platform Admin

Optional:

```text
PLATFORM_ADMIN
```

should be treated separately from workspace membership.

Its purpose:

```text
system operations

user status management

system diagnostics
```

It should not automatically mean:

```text
browse all private finance data
```

---

# 13. Resource Scope

P0 financial resources should be scoped by:

```text
workspace_id
```

rather than only:

```text
user_id
```

where feasible.

Examples:

```text
accounts.workspace_id

transactions.workspace_id

budgets.workspace_id

goals.workspace_id

scenarios.workspace_id
```

Audit fields still record:

```text
created_by_user_id
```

or actor identity where useful.

---

# 14. Personal Finance Still Remains Simple

The workspace model must remain invisible enough that a normal user still experiences:

```text
personal finance
```

not:

```text
enterprise workspace management
```

A default personal workspace can be automatically selected.

Workspace switching may be minimal if only one workspace exists.

---

# 15. Important Scope Correction — Account Balance

The authoritative account model should prioritize financial integrity.

Preferred final rule:

> **Account balance is derived from opening balance plus posted financial effects.**

Do not make a mutable `current_balance` field the ultimate source of truth.

---

# 16. Balance Implementation Options

Recommended P0 implementation:

```text
accounts.opening_balance
+
transactions
+
transfers
```

with account balance calculated through optimized aggregate queries.

If performance optimization later justifies cached balance:

```text
current_balance_cache
```

may exist as derived state.

It must never become independent financial truth.

---

# 17. Why Derived Balance Is Preferred

Benefits:

```text
financial history remains authoritative

reversal becomes easier to reason about

reconciliation is represented as adjustment

balance can be reconstructed

reduced risk of hidden drift
```

For personal-finance take-home volume, aggregate balance queries are acceptable.

---

# 18. Account Reconciliation

Reconciliation should create:

```text
ADJUSTMENT
```

financial record.

Example:

```text
Calculated balance:
Rp4.8M

Actual balance:
Rp5M

Adjustment:
+Rp200k
```

The application must not directly overwrite historical truth.

---

# 19. Implementation Milestone Overview

Recommended milestone sequence:

```text
M00 — Repository & Documentation Bootstrap

M01 — Infrastructure & Developer Experience

M02 — Backend Foundation

M03 — Database Foundation & Migrations

M04 — Authentication & Session Security

M05 — Workspace Authorization & RBAC

M06 — Frontend Foundation & Authentication

M07 — Accounts & Categories

M08 — Transactions & Financial Ledger

M09 — Transfers & Reconciliation

M10 — Recurring Transactions

M11 — Analytics & Dashboard Core

M12 — Budgets

M13 — Basic Financial Goals

M14 — Forecast Engine

M15 — Scenario Simulator

M16 — AI Foundation

M17 — AI Categorization & Insights

M18 — Savio Copilot

M19 — Frontend Intelligence Experience

M20 — Background Jobs & Notifications

M21 — Security Hardening

M22 — Testing Hardening

M23 — Performance & Observability

M24 — Documentation & API Finalization

M25 — Final UX Polish

M26 — Submission Audit
```

---

# 20. Milestone Dependency Graph

```text
M00
 ↓
M01
 ↓
M02
 ↓
M03
 ↓
M04
 ↓
M05
 ↓
M06
 ↓
M07
 ↓
M08
 ├─────────────┐
 ↓             ↓
M09           M11
 ↓             ↓
M10           M12
               ↓
              M13
               ↓
              M14
               ↓
              M15
               ↓
              M16
               ↓
              M17
               ↓
              M18
               ↓
              M19
               ↓
              M20
               ↓
              M21
               ↓
              M22
               ↓
              M23
               ↓
              M24
               ↓
              M25
               ↓
              M26
```

Some milestones may overlap, but dependency correctness must be preserved.

---

# 21. Milestone M00 — Repository & Documentation Bootstrap

## Objective

Create the repository structure and establish documentation as source of truth.

---

## Tasks

Create:

```text
savio/
├── README.md
├── AGENTS.md
├── DESIGN.md
├── backend/
├── frontend/
├── docs/
├── docker/
├── scripts/
├── .github/
├── docker-compose.yml
├── Makefile
└── .env.example
```

Create documentation structure:

```text
docs/
├── product/
│   ├── product-foundation.md
│   ├── business-requirements.md
│   └── user-flows.md
│
├── architecture/
│   ├── system-architecture.md
│   ├── frontend-architecture.md
│   └── ai-architecture.md
│
├── database/
│   └── database-design.md
│
├── api/
│   ├── api-contract.md
│   └── openapi.yaml
│
├── engineering/
│   ├── security.md
│   ├── testing-strategy.md
│   └── implementation-plan.md
│
└── assignment/
    └── take-home-test-specification.md
```

---

## Deliverables

```text
repository structure committed

documentation present

basic .gitignore

.env.example

initial README placeholder
```

---

## Definition of Done

```text
[ ] Repository initialized

[ ] Documentation tree exists

[ ] No secrets committed

[ ] Project name consistently Savio

[ ] Initial commit clean
```

---

# 22. Milestone M01 — Infrastructure & Developer Experience

## Objective

Make Savio reproducible locally.

---

## Infrastructure

Provision:

```text
PostgreSQL

Redis

MinIO
```

through:

```text
docker-compose.yml
```

---

## Docker Services

Recommended:

```text
postgres

redis

minio

minio-init
```

API/frontend containers may be added later.

---

## Environment Variables

Create:

```env
APP_ENV=

APP_PORT=

DATABASE_URL=

REDIS_URL=

JWT_SECRET=

CSRF_SECRET=

ACCESS_TOKEN_TTL=

REFRESH_TOKEN_TTL=

FRONTEND_ORIGIN=

AI_ENABLED=

AI_PROVIDER=

AI_BASE_URL=

AI_API_KEY=

AI_MODEL=

MINIO_ENDPOINT=

MINIO_ACCESS_KEY=

MINIO_SECRET_KEY=

MINIO_BUCKET=
```

---

## Makefile

Initial commands:

```bash
make infra-up

make infra-down

make migrate-up

make migrate-down

make test

make dev-api

make dev-worker

make dev-web
```

Commands may initially be placeholders until respective components exist.

---

## Deliverables

```text
Docker Compose starts infrastructure

PostgreSQL reachable

Redis reachable

MinIO reachable

.env.example complete
```

---

## Definition of Done

```text
[ ] docker compose up succeeds

[ ] PostgreSQL connection tested

[ ] Redis connection tested

[ ] MinIO initialized

[ ] No service credentials hardcoded

[ ] README local infrastructure section started
```

---

# 23. Milestone M02 — Backend Foundation

## Objective

Bootstrap the Go backend with a maintainable modular architecture.

---

## Initialize Backend

```bash
go mod init ...
```

Dependencies:

```text
Gin

GORM

PostgreSQL driver

UUID

decimal library

validator

structured logger

migration tooling
```

Potential:

```text
github.com/gin-gonic/gin

gorm.io/gorm

gorm.io/driver/postgres

github.com/google/uuid

github.com/shopspring/decimal
```

Exact package choices may vary.

---

## Backend Structure

```text
backend/
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── worker/
│       └── main.go
│
├── internal/
│   ├── platform/
│   └── ...
│
├── migrations/
├── seeds/
├── tests/
├── go.mod
└── go.sum
```

---

## Platform Components

Implement:

```text
config

database connection

logger

request ID

central error model

response helper

validator

health endpoint
```

---

## Initial Endpoints

```text
GET /health

GET /ready
```

---

## Middleware

Initial:

```text
Request ID

Recovery

Logging

CORS

Security headers
```

Authentication comes later.

---

## Deliverables

```text
Go API boots

structured logging

central errors

health endpoint

readiness endpoint

graceful shutdown
```

---

## Definition of Done

```text
[ ] go build succeeds

[ ] go test ./... succeeds

[ ] /health works

[ ] /ready checks PostgreSQL

[ ] Request ID returned

[ ] Panic does not expose stack trace

[ ] Graceful shutdown works
```

---

# 24. Milestone M03 — Database Foundation & Migrations

## Objective

Implement reproducible PostgreSQL schema foundation.

---

## Migration Tool

Use:

```text
golang-migrate
```

or equivalent explicit SQL migration system.

Do not use production `AutoMigrate` as source of truth.

---

## Initial Schema

Prioritize:

```text
users

workspaces

workspace_memberships

auth_sessions

user_settings

accounts

categories
```

Later migrations add feature tables.

---

## Workspace Tables

Recommended:

```text
workspaces

workspace_memberships
```

Concept:

```sql
workspaces
-----------
id
name
base_currency
timezone
created_by
created_at
updated_at
```

Membership:

```sql
workspace_memberships
---------------------
id
workspace_id
user_id
role
status
created_at
updated_at
```

Unique:

```text
workspace_id + user_id
```

---

## Workspace Roles

Allowed:

```text
OWNER

MEMBER

VIEWER
```

---

## Migration Requirements

Every migration:

```text
*.up.sql

*.down.sql
```

---

## Seed Strategy

Create:

```text
system categories

demo dataset
```

separately.

---

## Deliverables

```text
empty DB → latest schema

rollback available

constraints present

indexes intentional
```

---

## Definition of Done

```text
[ ] migrate up from empty DB succeeds

[ ] migrate down succeeds

[ ] migration test exists

[ ] no schema dependency on AutoMigrate

[ ] system category seed works

[ ] schema documented
```

---

# 25. Milestone M04 — Authentication & Session Security

## Objective

Implement secure cookie-based authentication.

---

## Required Endpoints

```text
GET  /api/v1/auth/csrf

POST /api/v1/auth/register

POST /api/v1/auth/login

POST /api/v1/auth/refresh

POST /api/v1/auth/logout

POST /api/v1/auth/logout-all

GET  /api/v1/auth/me
```

---

## Registration Transaction

Registration should atomically create:

```text
user

default workspace

OWNER membership

settings
```

Then auth session can be created.

---

## Password Security

Implement:

```text
Argon2id
```

or:

```text
bcrypt
```

No plaintext storage.

---

## Access Token

Implement:

```text
short-lived JWT
```

stored in:

```text
HttpOnly cookie
```

---

## Refresh Token

Implement:

```text
opaque random token
```

Database stores:

```text
hash only
```

---

## Refresh Rotation

Implement:

```text
row-locked session rotation
```

---

## CSRF

Implement:

```text
CSRF bootstrap

CSRF cookie

X-CSRF-Token validation
```

Protect:

```text
POST

PUT

PATCH

DELETE
```

according to security architecture.

---

## Login Protection

Implement:

```text
rate limit

generic invalid-credential response
```

---

## Deliverables

```text
secure login

secure refresh

secure logout

CSRF

session persistence
```

---

## Required Tests

```text
registration

duplicate email

password hashing

login success

invalid credentials

access expiry

refresh

refresh rotation

refresh replay rejection

concurrent refresh

logout

logout-all

CSRF valid

CSRF invalid
```

---

## Definition of Done

```text
[ ] Auth tokens never returned for localStorage use

[ ] Access cookie HttpOnly

[ ] Refresh cookie HttpOnly

[ ] Refresh token hashed

[ ] Rotation atomic

[ ] CSRF required

[ ] Rate limit active

[ ] Auth tests pass
```

---

# 26. Milestone M05 — Workspace Authorization & RBAC

## Objective

Implement meaningful multi-level backend authorization.

---

## Authorization Context

After authentication, request context should include:

```text
user_id

active_workspace_id

workspace_role
```

---

## Active Workspace

For MVP, default workspace may be selected automatically.

Potential request mechanism:

```text
X-Workspace-ID
```

or server-selected active workspace.

Recommended simple approach:

```text
single/default workspace automatically active
```

with workspace ID included in internal authenticated context.

If multi-workspace support is exposed later, add explicit workspace switching.

---

## Permission Matrix

Initial:

| Capability | OWNER | MEMBER | VIEWER |
| --- | --- | --- | --- |
| View Accounts | Yes | Yes | Yes |
| Create Account | Yes | Yes | No |
| Edit Account | Yes | Yes | No |
| Create Transaction | Yes | Yes | No |
| Edit Transaction | Yes | Yes | No |
| View Analytics | Yes | Yes | Yes |
| Create Budget | Yes | Yes | No |
| Create Scenario | Yes | Yes | No |
| Use AI Copilot | Yes | Yes | Yes |
| Manage Members | Yes | No | No |
| Change Roles | Yes | No | No |
| Delete Workspace | Yes | No | No |

---

## Last Owner Rule

Must prevent:

```text
removing last OWNER
```

or:

```text
demoting last OWNER
```

without ownership transfer.

---

## Member Endpoints

Minimal P0:

```text
GET  /api/v1/workspaces/current

GET  /api/v1/workspaces/current/members

POST /api/v1/workspaces/current/members

PATCH /api/v1/workspaces/current/members/:memberId

DELETE /api/v1/workspaces/current/members/:memberId
```

Invitation implementation may be simplified.

---

## Invitation Scope

Simplest defensible implementation:

```text
OWNER adds an existing Savio user by email
```

rather than implementing full email invitation infrastructure.

If email does not exist:

```text
return a clear validation error
```

Future email invitation can remain P1.

---

## Required Tests

```text
OWNER full access

MEMBER write access

VIEWER read-only

foreign workspace denied

last owner cannot be removed

member role update rules
```

---

## Definition of Done

```text
[ ] More than one meaningful role exists

[ ] Backend enforces roles

[ ] Frontend hiding buttons is not sole protection

[ ] Resources workspace-scoped

[ ] Cross-workspace IDOR tests pass

[ ] Last OWNER protected
```

---

# 27. Milestone M06 — Frontend Foundation & Authentication

## Objective

Bootstrap the React frontend and secure authentication behavior.

---

## Initialize Frontend

```text
React

TypeScript

Vite

Tailwind

React Router

Axios

TanStack Query

React Hook Form

Zod
```

---

## Create Application Structure

Implement:

```text
app/

features/

shared/
```

according to frontend architecture.

---

## Initial Layouts

```text
AuthLayout

AppLayout
```

---

## Initial Routes

```text
/login

/register

/dashboard
```

---

## Auth Bootstrap

State:

```text
UNKNOWN

AUTHENTICATED

UNAUTHENTICATED
```

Use:

```text
GET /auth/me
```

---

## Axios Client

Implement:

```text
withCredentials

CSRF header

401 handling

single-flight refresh

403 handling

422 mapping

429 handling

500 safe errors
```

---

## Query Cache Logout

On logout:

```text
clear TanStack Query cache
```

---

## Deliverables

```text
login UI

register UI

protected routes

auth bootstrap

refresh interceptor
```

---

## Required Tests

```text
protected route

guest route

login validation

single-flight refresh

refresh failure

no infinite loop

logout cache clear
```

---

## Definition of Done

```text
[ ] No tokens in localStorage

[ ] No tokens in sessionStorage

[ ] Protected routes work

[ ] Refresh single-flight works

[ ] 403 does not logout

[ ] 422 maps to form

[ ] Auth UI responsive
```

---

# 28. Milestone M07 — Accounts & Categories

## Objective

Implement the first authoritative financial resources.

---

## Backend Modules

```text
accounts

categories
```

---

## Account Fields

Core:

```text
workspace_id

name

type

currency

opening_balance

institution_name

description

status

version
```

---

## Account Balance

Expose:

```text
derived_balance
```

calculated from:

```text
opening balance

posted income

posted expense

adjustments

transfers
```

---

## Account Endpoints

```text
GET    /accounts

POST   /accounts

GET    /accounts/:id

PATCH  /accounts/:id

POST   /accounts/:id/archive

POST   /accounts/:id/restore

DELETE /accounts/:id
```

Delete only when safe.

---

## Category Endpoints

```text
GET    /categories

POST   /categories

PATCH  /categories/:id

POST   /categories/:id/archive

POST   /categories/:id/restore
```

---

## Category Rules

```text
INCOME categories

EXPENSE categories

system categories

workspace/user custom categories
```

---

## Frontend

Build:

```text
Accounts Page

Account Card

Create Account Form

Edit Account Form

Categories management if exposed
```

---

## Required Tests

```text
account ownership

workspace authorization

currency rule

archive rule

category type

system category access

custom category isolation
```

---

## Definition of Done

```text
[ ] Account CRUD business rules work

[ ] Derived balance endpoint ready

[ ] Categories seeded

[ ] Archived account cannot accept new finance writes

[ ] VIEWER cannot mutate
```

---

# 29. Milestone M08 — Transactions & Financial Ledger

## Objective

Implement the authoritative financial ledger behavior.

This is one of the most critical milestones.

---

## Transaction Types

```text
INCOME

EXPENSE

ADJUSTMENT
```

Transfer is handled separately.

---

## Recommended Transaction Lifecycle

Use:

```text
POSTED

REVERSED
```

rather than overbuilding a draft workflow.

A transaction becomes authoritative when created.

Corrections may use:

```text
edit with audit + optimistic version
```

or:

```text
reverse + replacement
```

Recommended P0 balance:

> Allow edits while preserving audit history and correctly recalculating ledger effects.

Reversal remains available for explicit invalidation.

---

## Transaction Fields

```text
workspace_id

account_id

category_id

type

amount

transaction_date

description

merchant

notes

source

status

version

created_by_user_id

created_at

updated_at
```

---

## Money Representation

Use:

```text
PostgreSQL NUMERIC
+
decimal-safe Go type
```

API:

```text
decimal string
```

---

## Transaction Endpoints

```text
GET  /transactions

POST /transactions

GET  /transactions/:id

PATCH /transactions/:id

POST /transactions/:id/reverse
```

---

## List Capabilities

Implement:

```text
search

type filter

account filter

category filter

date range

status

sort

pagination
```

---

## Financial Rules

INCOME:

```text
category type = INCOME
```

EXPENSE:

```text
category type = EXPENSE
```

ADJUSTMENT:

```text
created through reconciliation / controlled flow
```

---

## Analytics Exclusion

Only:

```text
POSTED
```

transactions count.

Adjustments should be handled separately in income/expense analytics unless explicitly intended.

Recommended:

```text
ADJUSTMENT excluded from ordinary income/expense analytics
```

---

## Audit

Record:

```text
create

update

reverse
```

---

## Optimistic Locking

Transaction updates require:

```text
version
```

---

## Frontend

Build:

```text
Transactions Page

Search

Filters

Sorting

Pagination

Add Income

Add Expense

Edit Transaction

Transaction Detail

Reverse Confirmation
```

---

## Required Tests

```text
create income

create expense

category mismatch

foreign account

archived account

update

version conflict

reverse

reversed analytics exclusion

workspace isolation

search isolation

pagination
```

---

## Definition of Done

```text
[ ] Financial ledger authoritative

[ ] Amount precision exact

[ ] Search/filter/sort/pagination work

[ ] Audit events created

[ ] Version conflicts return 409

[ ] Reversal works

[ ] Transaction page polished
```

---

# 30. Milestone M09 — Transfers & Reconciliation

## Objective

Implement money movement between internal accounts and account reconciliation.

---

## Transfer Model

Preferred representation:

```text
transfer domain record
+
source account effect
+
destination account effect
```

Implementation may use:

```text
two linked ledger entries
```

or calculate transfers separately.

The key invariant:

> Transfers must never count as income or expense.

---

## Transfer Endpoints

```text
GET  /transfers

POST /transfers

GET  /transfers/:id

POST /transfers/:id/reverse
```

---

## Rules

```text
source != destination

same workspace

same supported currency

both accounts active

amount > 0
```

---

## Atomicity

Transfer creation/reversal must be transactional.

---

## Reconciliation Endpoint

```text
POST /accounts/:id/reconcile
```

Request:

```text
actual_balance

reason

version
```

Backend calculates:

```text
difference
```

and creates:

```text
ADJUSTMENT
```

---

## Frontend

Build:

```text
Transfer Form

Transfer Preview

Transfer Detail

Reverse Transfer

Reconcile Account Modal
```

---

## Required Tests

```text
transfer invariant

same account rejected

cross-workspace rejected

reversal

reconciliation positive

reconciliation negative

financial analytics unchanged
```

---

## Definition of Done

```text
[ ] Transfer does not affect income/expense

[ ] Reconciliation preserves history

[ ] Cross-account flow polished

[ ] Atomicity tests pass
```

---

# 31. Milestone M10 — Recurring Transactions

## Objective

Model future known financial activity.

---

## Recurring Types

```text
INCOME

EXPENSE
```

---

## Status

Recommended:

```text
ACTIVE

PAUSED

ENDED
```

If existing implementation/docs use:

```text
COMPLETED

CANCELLED
```

choose one final vocabulary and normalize the whole codebase.

Recommended final product vocabulary:

```text
ACTIVE

PAUSED

ENDED
```

because it is simpler.

---

## Important Behavior

Recurring rules should generate:

```text
scheduled/projected occurrences
```

but should not automatically become actual financial history by default.

Recommended P0:

```text
auto_post = false by default
```

User confirms an occurrence as actual.

---

## Why Confirmation by Default

This avoids assuming:

```text
a planned bill was actually paid

a planned salary was actually received
```

It also aligns with:

> **User decides.**

---

## Optional Auto-Post

If implemented:

```text
auto_post=true
```

must be explicit and idempotent.

---

## Endpoints

```text
GET   /recurring-transactions

POST  /recurring-transactions

GET   /recurring-transactions/:id

PATCH /recurring-transactions/:id

POST /recurring-transactions/:id/pause

POST /recurring-transactions/:id/resume

POST /recurring-transactions/:id/end

GET  /recurring-transactions/:id/occurrences

POST /recurring-occurrences/:id/confirm

POST /recurring-occurrences/:id/skip
```

---

## Occurrence Status

Recommended:

```text
PLANNED

CONFIRMED

SKIPPED
```

Optional auto-post failure state:

```text
FAILED
```

---

## Recurring Date Logic

Test:

```text
daily

weekly

monthly

month-end

leap year

end date
```

---

## Frontend

Build:

```text
Recurring List

Recurring Form

Upcoming Occurrences

Confirm Occurrence

Skip Occurrence

Pause / Resume
```

---

## Required Tests

```text
date generation

duplicate occurrence prevention

confirmation idempotency

pause

resume

end

month end
```

---

## Definition of Done

```text
[ ] Future recurring activity appears in forecast

[ ] Actual ledger changes only when confirmed/auto-posted

[ ] Duplicate posting impossible

[ ] UI distinguishes planned vs actual
```

---

# 32. Milestone M11 — Analytics & Dashboard Core

## Objective

Turn ledger data into understandable financial summaries.

---

## Analytics Service

Implement deterministic:

```text
income

expense

net cashflow

savings rate

category breakdown

period comparison

spending changes

recurring expense summary
```

---

## Rules

Exclude:

```text
transfers

reversed transactions

adjustments from normal income/expense metrics
```

unless explicitly requested.

---

## Endpoints

```text
GET /analytics/cashflow

GET /analytics/categories

GET /analytics/period-comparison

GET /analytics/recurring-expenses

GET /analytics/spending-changes

GET /dashboard
```

---

## Dashboard Response

Include:

```text
balance summary

cashflow

budgets preview later

goal preview later

forecast preview later

insights preview later

upcoming recurring events
```

Missing modules may initially return empty/omitted sections.

---

## Database

Use:

```text
aggregate SQL queries
```

rather than loading all transactions into application memory.

---

## Frontend

Build:

```text
Dashboard Skeleton

Financial Summary Cards

Cashflow Chart

Analytics Page

Category Breakdown

Period Comparison
```

---

## Required Tests

```text
income

expense

transfer exclusion

reversal exclusion

adjustment exclusion

period comparison

cross-workspace isolation
```

---

## Definition of Done

```text
[ ] Dashboard answers current financial position

[ ] Analytics are deterministic

[ ] DB aggregation used

[ ] No duplicated financial formulas in frontend
```

---

# 33. Milestone M12 — Budgets

## Objective

Implement category-based spending planning.

---

## P0 Budget Scope

Keep budgets intentionally simple:

```text
monthly

expense category

one active budget/category/period
```

Avoid arbitrary complex recurring budget schedules initially.

---

## Fields

```text
workspace_id

category_id

amount

period_start

period_end

status

version
```

---

## Status

Recommended:

```text
ACTIVE

CLOSED
```

Historical periods naturally become closed/read-only if desired.

---

## Derived Values

Backend calculates:

```text
spent

remaining

utilization

status

projected spend

projected overspend
```

---

## Computed Budget Status

Separate persistence status from computed financial condition.

Persistence:

```text
ACTIVE
CLOSED
```

Computed:

```text
ON_TRACK
WARNING
EXCEEDED
```

This distinction avoids overloading one status field.

---

## Endpoints

```text
GET  /budgets

POST /budgets

GET  /budgets/:id

PATCH /budgets/:id

POST /budgets/:id/close
```

---

## Frontend

Build:

```text
Budget List

Budget Progress

Create Budget

Edit Budget

Budget Detail

Projected Risk
```

---

## Required Tests

```text
monthly spend

reversed expense exclusion

transfer exclusion

duplicate budget

warning threshold

exceeded

version conflict
```

---

## Definition of Done

```text
[ ] Budget actual is derived

[ ] Projected vs actual clearly separated

[ ] Duplicate budget prevented

[ ] UI exposes deterministic status
```

---

# 34. Milestone M13 — Basic Financial Goals

## Objective

Add lightweight goal planning without overbuilding financial allocation accounting.

---

## Scope Decision

Goals are useful, but not central enough to justify complex allocation ledgers during the take-home.

Recommended P0 goal model:

```text
target amount

current planned amount

target date

priority
```

The user explicitly maintains `current_amount`.

Clearly label it as:

```text
tracked goal progress
```

not an automatically reserved bank balance.

---

## Optional Linked Account

If implemented:

```text
linked_account_id
```

is informational.

Do not assume:

```text
whole account balance = goal money
```

---

## Goal Status

```text
ACTIVE

PAUSED

ACHIEVED

CANCELLED
```

---

## Deterministic Metrics

```text
progress

remaining

months remaining

required monthly contribution

simple feasibility
```

---

## Feasibility

Example:

```text
required monthly contribution
vs
estimated free cashflow
```

Output:

```text
ON_TRACK

AT_RISK
```

Do not overclaim full financial health.

---

## Frontend

Build:

```text
Goal Cards

Goal Form

Goal Detail

Progress

Feasibility
```

---

## Required Tests

```text
progress

100% cap display

required contribution

target-date edge

achieved
```

---

## Definition of Done

```text
[ ] Goal semantics clearly documented

[ ] No false account allocation claim

[ ] Goal metrics deterministic

[ ] Goal UI useful but lightweight
```

---

# 35. Milestone M14 — Forecast Engine

## Objective

Implement Savio's first major differentiated capability.

---

## Forecast Principle

Forecast is:

```text
deterministic
+
assumption-driven
+
explainable
```

It is not:

```text
LLM prediction
```

---

## Default Horizon

Recommended:

```text
90 days
```

Supported:

```text
30

60

90

180

365
```

---

## Forecast Inputs

```text
derived current liquid balance

active recurring income

active recurring expense

planned recurring occurrences

historical variable expenses

user-defined assumptions if present
```

---

## Historical Baseline

Recommended initial variable-spending method:

```text
trailing 90-day average
```

If less history:

```text
use available history

flag limited confidence
```

---

## Minimum History

Potential confidence rules:

```text
< 30 days
→ LOW

30–89 days
→ MEDIUM

>= 90 days
→ HIGH
```

These are initial heuristics.

Confidence may also consider known recurring coverage.

Keep methodology deterministic.

---

## Forecast Outputs

```text
opening balance

ending balance

minimum balance

minimum-balance date

projected income

projected expense

timeline

assumptions

confidence

calculation version
```

---

## Event Types

```text
CONFIRMED

RECURRING

ESTIMATED

ASSUMED
```

User-facing labels should match design documentation.

---

## Snapshot Persistence

Implement:

```text
optional persisted snapshot
```

for:

```text
history

scenario baseline

reproducibility
```

---

## Freshness

Financial mutations mark relevant snapshots:

```text
STALE
```

---

## Frontend

Build:

```text
Forecast Page

Horizon Selector

Summary Cards

Balance Chart

Timeline

Assumptions

Confidence

Stale Banner
```

---

## Required Tests

```text
event ordering

minimum balance

ending balance

recurring event

estimated spend

insufficient history

confidence

transfer exclusion

snapshot

stale state

timezone
```

---

## Definition of Done

```text
[ ] Forecast fully deterministic

[ ] No AI dependency

[ ] Assumptions visible

[ ] Confidence explainable

[ ] Timeline explainable

[ ] Tests cover edge cases
```

---

# 36. Milestone M15 — Scenario Simulator

## Objective

Implement Savio's primary killer feature.

---

## Scenario Principle

A scenario is:

> **A non-destructive overlay applied to the forecast baseline.**

It must never modify real financial data.

---

## Scenario Types

P0:

```text
ONE_TIME_EXPENSE

ONE_TIME_INCOME

RECURRING_EXPENSE

RECURRING_INCOME

INCOME_REDUCTION

INCOME_REMOVAL

EXPENSE_REDUCTION
```

Optional:

```text
SAVINGS_ADJUSTMENT
```

---

## Scenario Entities

```text
financial_scenarios

scenario_modifications

scenario_snapshots
```

---

## Scenario Flow

```text
Load authoritative data
↓
Build baseline
↓
Apply modifications
↓
Recalculate projection
↓
Compare
↓
Persist snapshot
```

---

## Outputs

```text
baseline ending balance

scenario ending balance

baseline minimum balance

scenario minimum balance

cashflow difference

savings-rate difference

cash-runway difference

goal impact

assumptions
```

---

## Stale Detection

Financial source changes:

```text
scenario.is_stale = true
```

---

## Endpoints

```text
GET    /scenarios

POST   /scenarios

GET    /scenarios/:id

PATCH  /scenarios/:id

DELETE /scenarios/:id

POST   /scenarios/:id/modifications

PATCH  /scenarios/:id/modifications/:id

DELETE /scenarios/:id/modifications/:id

POST   /scenarios/:id/calculate

GET    /scenarios/:id/snapshots
```

---

## Frontend

Build polished:

```text
Scenario Builder

Modification Cards

Calculate Action

Baseline vs Scenario

Difference Metrics

Projection Overlay Chart

Goal Impact

Assumptions

Stale State
```

---

## Required Tests

```text
one-time purchase

income reduction

income removal

recurring expense

multiple modifications

baseline isolation

real DB state unchanged

snapshot

goal impact

stale state
```

---

## Definition of Done

```text
[ ] Real financial state never mutated

[ ] Scenario deterministic

[ ] Comparison understandable

[ ] Killer feature demo-ready

[ ] Tests prove isolation
```

---

# 37. Milestone M16 — AI Foundation

## Objective

Introduce AI only after deterministic financial intelligence exists.

---

## AI Module

Implement:

```text
provider interface

mock provider

real provider adapter

context builder

privacy filter

prompt registry

output schemas

orchestrator

request logging
```

---

## Provider Interface

Support:

```text
OpenAI-compatible provider
```

plus:

```text
mock provider
```

---

## Environment

```env
AI_ENABLED=true

AI_PROVIDER=openai_compatible

AI_BASE_URL=

AI_API_KEY=

AI_MODEL=

AI_TIMEOUT_SECONDS=20
```

---

## AI Failure

Core app must operate if:

```text
AI_ENABLED=false
```

or provider fails.

---

## Structured Output

Implement strict schemas for:

```text
categorization

insight

copilot

scenario explanation
```

---

## Tool Registry

Initial tools:

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

---

## Security

No generic:

```text
SQL

HTTP

shell

filesystem
```

tools.

---

## Required Tests

```text
mock provider

timeout

provider failure

invalid output

tool validation

cross-workspace protection

context minimization
```

---

## Definition of Done

```text
[ ] Provider swappable

[ ] Mock works

[ ] AI key server-only

[ ] Finance has no AI dependency

[ ] Output validated

[ ] Tools bounded
```

---

# 38. Milestone M17 — AI Categorization & Insights

## Objective

Implement the first visible AI-assisted finance experiences.

---

## Transaction Categorization

Flow:

```text
description / merchant
↓
available category keys
↓
AI
↓
validated suggestion
↓
user confirms
```

AI does not save the transaction itself.

---

## Categorization Result

```text
category

confidence if meaningful

reason
```

---

## Insight Architecture

Use:

```text
deterministic signal
↓
AI explanation
```

not:

```text
dump entire ledger into LLM
```

---

## Initial Signals

P0:

```text
SPENDING_ANOMALY

BUDGET_RISK

CASHFLOW_RISK

INCOME_CHANGE

POSITIVE_TREND
```

---

## Signal Detection

Deterministic rules.

Example:

```text
category spending
> baseline by threshold
```

---

## AI Output

AI explains:

```text
what changed

main driver

what to review
```

Severity remains deterministic.

---

## Frontend

Build:

```text
AI Category Suggestion

Insights Feed

Insight Card

Insight Detail

Dismiss
```

---

## Required Tests

```text
suggestion valid

invalid category

signal severity preserved

dedup

provider failure

prompt injection

other-workspace data inaccessible
```

---

## Definition of Done

```text
[ ] AI is clearly a suggestion

[ ] Insight facts shown before narrative

[ ] AI failure non-blocking

[ ] Prompt injection test passes
```

---

# 39. Milestone M18 — Savio Copilot

## Objective

Implement grounded natural-language access to Savio financial intelligence.

---

## P0 Scope

Copilot supports:

```text
READ

ANALYZE

COMPARE

SIMULATE

EXPLAIN
```

Do not implement broad autonomous mutation.

---

## Supported Questions

```text
Why did I spend more this month?

Where did my money go?

What are my largest recurring expenses?

Which budget is at risk?

What does my forecast look like?

What happens if I buy a Rp15M laptop?

What happens if my income drops?

Am I on track for my goal?
```

---

## Flow

```text
User Question
↓
Intent
↓
Required Tools
↓
Deterministic Finance Results
↓
Minimal Context
↓
LLM
↓
Structured Response
```

---

## Clarification

If required information missing:

```text
CLARIFICATION_REQUIRED
```

Example:

```text
purchase date missing
```

Do not guess unnecessarily.

---

## Copilot Response

Return:

```text
answer

facts

actions

sources/tool names

clarification if needed
```

---

## Scenario Handoff

Question:

```text
Can I afford a Rp15M laptop?
```

may use:

```text
temporary deterministic scenario
```

with option:

```text
Save as Scenario
```

---

## Frontend

Build:

```text
Copilot Page

Suggested Prompts

Message Composer

Fact Cards

Clarification Buttons

Actions

Error State

Rate-Limit State
```

---

## Required Tests

```text
intent

tool execution

fact grounding

clarification

AI unavailable

rate limit

foreign resource ID

unknown tool

unknown action
```

---

## Definition of Done

```text
[ ] Copilot grounded in tools

[ ] No direct DB access

[ ] No arbitrary actions

[ ] Missing data produces clarification

[ ] Numbers come from finance results
```

---

# 40. Milestone M19 — Frontend Intelligence Experience

## Objective

Bring Dashboard, Forecast, Scenario, Insights, and Copilot into one polished product experience.

---

## Dashboard Finalization

Add:

```text
budget summary

goal summary

forecast preview

insight preview

upcoming financial activity
```

---

## Forecast Polish

Add:

```text
historical vs projected visual distinction

today marker

event source labels

confidence explanation

low-balance warning
```

---

## Scenario Polish

Add:

```text
sticky builder if useful

difference emphasis

baseline/scenario chart

goal impact

AI explanation
```

---

## AI UX

Ensure:

```text
facts first

AI explanation second

actions third
```

---

## Responsive

Verify:

```text
375

768

1024

1440
```

---

## Accessibility

Verify:

```text
focus

labels

keyboard

dialogs

chart summaries

non-color statuses
```

---

## Definition of Done

```text
[ ] Core product journey visually coherent

[ ] Mobile usable

[ ] No generic admin-template feel

[ ] Forecast/scenario clearly differentiated from actual data

[ ] AI clearly differentiated from deterministic facts
```

---

# 41. Milestone M20 — Background Jobs & Notifications

## Objective

Move appropriate non-critical processing out of synchronous request flow.

---

## Worker

Implement:

```text
backend/cmd/worker
```

---

## Redis Queue

Potential:

```text
Asynq
```

or equivalent.

---

## Initial Jobs

```text
recurring occurrence generation

optional auto-post

financial signal evaluation

AI insight generation

notification generation

expired session cleanup
```

---

## Notification Types

P1:

```text
BUDGET_WARNING

LOW_PROJECTED_BALANCE

UPCOMING_RECURRING

AI_INSIGHT_AVAILABLE
```

---

## Idempotency

Every repeating job must be retry-safe.

---

## Important Rule

A financial write should not fail just because:

```text
AI insight enqueue failed
```

---

## Frontend

If notifications implemented:

```text
Bell

Unread Count

Notification Popover

Notification Page
```

---

## Required Tests

```text
job retry

deduplication

recurring idempotency

queue failure

AI job failure

notification dedup
```

---

## Definition of Done

```text
[ ] Worker shares domain services

[ ] No duplicate business logic

[ ] Jobs idempotent

[ ] Queue failure isolated
```

---

# 42. Milestone M21 — Security Hardening

## Objective

Perform focused security pass after feature integration.

---

## Review

Verify:

```text
cookie attributes

CSRF

CORS

RBAC

ownership

rate limiting

refresh rotation

cache clearing

error disclosure

security headers

trusted proxy

secrets

AI context

prompt injection
```

---

## Tests

Run:

```text
cross-workspace tests

CSRF tests

refresh race

rate limits

prompt injection

XSS rendering

invalid sort

invalid resource IDs
```

---

## Dependency Checks

Run:

```bash
govulncheck ./...
```

Frontend:

```bash
npm audit
```

or equivalent.

---

## Definition of Done

```text
[ ] Security checklist complete

[ ] No high-impact known vulnerability ignored

[ ] No secrets in Git history/current tree

[ ] Production config secure
```

---

# 43. Milestone M22 — Testing Hardening

## Objective

Complete automated coverage around highest-risk behavior.

---

## Mandatory Backend Tests

```text
Finance Engine

Budget Engine

Goal Engine

Forecast

Scenario

Transaction

Transfer

Recurring

Auth

CSRF

RBAC

IDOR

Concurrency
```

---

## Mandatory AI Tests

```text
mock provider

schema validation

tool authorization

provider failure

prompt injection
```

---

## Mandatory Frontend Tests

```text
auth bootstrap

401 single-flight refresh

422 field errors

409 conflict

AI degraded state

scenario comparison
```

---

## Mandatory E2E

```text
Register

Onboarding

Account

Income

Expense

Budget

Forecast

Scenario

AI Mock

Logout
```

---

## Race Detector

Run:

```bash
go test -race ./...
```

---

## Definition of Done

```text
[ ] No critical test skipped

[ ] CI green

[ ] E2E stable

[ ] Race detector reviewed

[ ] Financial invariants covered
```

---

# 44. Milestone M23 — Performance & Observability

## Objective

Ensure the project demonstrates engineering maturity without premature infrastructure complexity.

---

## Observability

Verify:

```text
structured logs

request ID

health

readiness

worker logs

AI request metrics/logging
```

---

## Performance Review

Check:

```text
N+1

pagination

indexes

analytics queries

forecast complexity
```

---

## Large Dataset

Seed:

```text
10,000 transactions
```

Review:

```text
transaction list

analytics

dashboard
```

---

## Query Analysis

Use:

```sql
EXPLAIN ANALYZE
```

for:

```text
transaction listing

category aggregation

recurring due query
```

---

## Potential Optimization

Only if needed:

```text
new indexes

query rewrite

bounded parallel reads
```

Do not add Redis caching without a demonstrated reason.

---

## Definition of Done

```text
[ ] No obvious N+1

[ ] Pagination bounded

[ ] Important indexes used

[ ] Logs useful

[ ] Readiness accurately models dependencies
```

---

# 45. Milestone M24 — Documentation & API Finalization

## Objective

Ensure repository explains itself to reviewers.

---

## OpenAPI

Finalize:

```text
docs/api/openapi.yaml
```

Cover:

```text
auth

CSRF

workspaces

accounts

transactions

transfers

recurring

budgets

goals

analytics

forecast

scenarios

AI

errors

pagination
```

---

## Swagger

Expose:

```text
/api/docs
```

if practical.

---

## README

Final README must contain:

```text
Project Overview

Problem

Target Users

Features

Product Thesis

Architecture

Tech Stack

Repository Structure

Prerequisites

Installation

Environment

Database Migration

Seed Data

Running Backend

Running Frontend

Running Worker

Testing

API Documentation

AI Configuration

Demo Credentials

Technical Decisions

Trade-offs

Future Improvements
```

---

## Documentation Consistency

Audit terminology:

```text
workspace vs user ownership

opening balance vs current balance

recurring status

transaction status

budget status

AI feature scope
```

Remove contradictions.

---

## Definition of Done

```text
[ ] README complete

[ ] OpenAPI accurate

[ ] Architecture docs align with code

[ ] Setup reproducible

[ ] Reviewer can run without asking questions
```

---

# 46. Milestone M25 — Final UX Polish

## Objective

Improve user-facing quality after functionality is stable.

---

## Review All Major Screens

```text
Login

Register

Onboarding

Dashboard

Accounts

Transactions

Recurring

Budgets

Goals

Analytics

Forecast

Scenarios

Insights

Copilot

Settings
```

---

## Verify States

Every important screen:

```text
loading

empty

success

error

pending

responsive
```

---

## Copy Review

Remove:

```text
Lorem ipsum

TODO text

technical backend wording

inconsistent capitalization
```

---

## Visual Review

Check:

```text
spacing

typography

button hierarchy

card consistency

chart labels

semantic statuses

AI distinction
```

---

## Mobile Review

Critical flow:

```text
login

add expense

forecast

scenario

Copilot
```

must be usable at mobile width.

---

## Definition of Done

```text
[ ] No broken responsive layouts

[ ] No placeholder UX

[ ] Critical flows polished

[ ] Accessibility basics verified

[ ] Visual hierarchy follows DESIGN.md
```

---

# 47. Milestone M26 — Submission Audit

## Objective

Perform final assessment-oriented review before delivery.

---

## Functional Audit

Verify:

```text
Auth

RBAC

CSRF

CRUD/business flows

search

filter

sort

pagination

forecast

scenario

AI

worker

Docker
```

---

## Backend Audit

Verify:

```text
handlers thin

services own business rules

repositories persistence-focused

finance engine deterministic

errors centralized

transactions explicit
```

---

## Frontend Audit

Verify:

```text
feature architecture

TanStack Query

forms

Axios interceptors

auth bootstrap

responsive

error states
```

---

## Security Audit

Verify:

```text
no localStorage auth

refresh rotation

CSRF

cookie flags

ownership

permissions

rate limiting

headers

safe errors
```

---

## Database Audit

Verify:

```text
migration from empty DB

rollback

constraints

indexes

FKs

integrity
```

---

## Testing Audit

Verify:

```text
backend tests

frontend tests

integration

security

concurrency

E2E
```

---

## Documentation Audit

Verify:

```text
README

API docs

architecture

security

testing

technical decisions

trade-offs
```

---

## Git Audit

Verify:

```text
meaningful commits

no giant "final" commit

no secrets

clean branches

clean status
```

---

# 48. Implementation Priority by Assessment Weight

The take-home evaluation should influence implementation order.

Recommended effort emphasis:

| Area | Weight / Importance | Implementation Priority |
| --- | ---: | --- |
| Backend Architecture | High | P0 |
| API & Business Logic | High | P0 |
| Authentication & Security | High | P0 |
| UI / UX | High | P0 |
| Database & Migrations | High | P0 |
| Frontend Architecture | High | P0 |
| Testing | Medium | P0 |
| Documentation | Medium | P0 |
| Git Quality | Medium | P0 |
| Bonus Features | Differentiator | After Core |

---

# 49. Features That Should Not Be Cut

If time is limited, preserve:

```text
secure cookie auth

CSRF

RBAC

accounts

transactions

transfers

recurring

budgets

analytics

forecast

scenario

at least one meaningful AI insight flow

Copilot grounded in tools

search/filter/sort/pagination

migrations

tests

responsive UI

README
```

---

# 50. Features That Can Be Simplified

If time becomes constrained:

```text
Goals
→ basic progress only

Notifications
→ omit or minimal

AI conversation persistence
→ stateless

Reports
→ omit

MinIO
→ infrastructure included, receipt feature omitted

Advanced admin
→ minimal operations page

Multiple workspaces UI
→ default personal workspace only

Workspace invite
→ add existing user by email only
```

---

# 51. Features That Can Be Removed First

Cut first:

```text
receipt OCR

CSV import

dark mode

advanced notifications

AI history

semantic search

advanced reports

provider fallback

streaming AI

household UX

multi-currency

bank sync
```

---

# 52. Scope Reduction Rule

When time is limited:

> **Reduce breadth before reducing correctness.**

Prefer:

```text
8 complete features
```

over:

```text
20 half-implemented features
```

---

# 53. Backend Module Implementation Pattern

For every backend module:

```text
1. Domain types

2. Repository interface

3. Repository implementation

4. Service

5. DTOs

6. Handler

7. Routes

8. Tests

9. OpenAPI update
```

---

# 54. Example — Transaction Module Flow

```text
domain/
transaction.go

repository/
transaction_repository.go

service/
transaction_service.go

handler/
transaction_handler.go

dto/
transaction_request.go
transaction_response.go

routes
```

Exact structure may be adjusted.

---

# 55. Frontend Feature Implementation Pattern

For every major frontend feature:

```text
1. API types

2. Query keys

3. API functions

4. Query hooks

5. Mutation hooks

6. Validation schema

7. Components

8. Page

9. Loading state

10. Empty state

11. Error state

12. Tests
```

---

# 56. Vertical Slice Strategy

After foundation, prefer vertical slices.

Example Transactions:

```text
migration
↓
repository
↓
service
↓
API
↓
frontend
↓
tests
```

rather than implementing all database tables first and UI much later.

---

# 57. Vertical Slice Benefits

```text
feedback earlier

API contracts validated

frontend/backend integration tested

less abandoned code

demo becomes usable incrementally
```

---

# 58. Recommended Git Commit Strategy

Commits should be meaningful.

Example:

```text
chore: bootstrap Savio repository

chore: add local infrastructure

feat(auth): add registration and cookie sessions

feat(auth): implement refresh token rotation

feat(security): add CSRF protection

feat(workspaces): add membership RBAC

feat(accounts): add account management

feat(transactions): implement financial ledger

feat(transfers): add atomic account transfers

feat(recurring): add recurring transaction planning

feat(analytics): add cashflow analytics

feat(forecast): implement deterministic cashflow forecast

feat(scenarios): add non-destructive scenario simulator

feat(ai): add provider abstraction

feat(ai): add transaction category suggestions

feat(ai): add grounded Savio Copilot

test: add concurrency and financial invariant coverage

docs: finalize architecture and setup guide
```

---

# 59. Git Rules

Avoid commits such as:

```text
update

fix

changes

final

test
```

without meaningful context.

---

# 60. Branch Strategy

For solo take-home:

```text
main
```

plus short-lived feature branches is sufficient.

Example:

```text
feat/auth

feat/transactions

feat/forecast
```

Avoid unnecessary Git-flow complexity.

---

# 61. Definition of Done — Backend Feature

A backend feature is complete when:

```text
migration exists if required

business logic implemented

authorization enforced

validation implemented

error mapping implemented

tests pass

OpenAPI updated

no obvious N+1
```

---

# 62. Definition of Done — Financial Feature

Additional requirements:

```text
financial invariant defined

atomicity verified

concurrency considered

audit event considered

forecast freshness considered

scenario freshness considered
```

---

# 63. Definition of Done — Frontend Feature

```text
API integrated

loading state

empty state

error state

form validation

backend validation mapping

responsive

keyboard basics

tests
```

---

# 64. Definition of Done — AI Feature

```text
deterministic input source

minimal context

provider abstraction

mock test

schema validation

error handling

rate limiting

authorization

user control
```

---

# 65. Definition of Done — Security Feature

```text
threat identified

server-side enforcement

negative test

failure response safe

no sensitive logging
```

---

# 66. Core Business Invariants to Protect During Coding

These invariants must remain visible throughout implementation.

```text
1. A user may access only authorized workspace data.

2. VIEWER cannot mutate finance state.

3. Money uses decimal-safe arithmetic.

4. Transfers do not count as income or expense.

5. Reversed transactions do not count in analytics.

6. Account balance is reconstructable from ledger state.

7. Reconciliation creates an adjustment.

8. Recurring plans do not silently become actual transactions unless confirmed/auto-posted.

9. A recurring occurrence is posted at most once.

10. Forecast is deterministic.

11. Scenario never mutates real finance data.

12. AI never becomes source of financial truth.

13. AI cannot choose authorization context.

14. Financial writes are backend validated.

15. Client-side calculations are never authoritative.
```

---

# 67. Architecture Guardrails During Implementation

Do not introduce:

```text
generic BaseService for every domain

generic CRUD controller

business logic in handlers

financial formulas in frontend

AI calls inside finance engine

raw SQL tools for AI

mutable balance without ledger reconciliation

production AutoMigrate

auth token localStorage
```

---

# 68. Backend Dependency Guardrail

Desired:

```text
Handler
↓
Service
↓
Domain / Finance Engine
↓
Repository
```

AI:

```text
AI Orchestrator
↓
Finance Service
```

Never:

```text
Finance Engine
↓
AI
```

---

# 69. Frontend Dependency Guardrail

Desired:

```text
Page
↓
Feature Hook
↓
Feature API
↓
Shared Axios Client
```

Avoid direct Axios calls scattered throughout components.

---

# 70. Database Guardrail

Before adding a new persisted field, ask:

```text
Is this authoritative?

Is it derived?

Can it become stale?

Can it be recalculated?
```

If derived:

```text
prefer calculation
```

unless persistence has a clear purpose.

---

# 71. AI Guardrail

Before sending a value to model, ask:

```text
Does AI actually need this?
```

If no:

```text
do not send it.
```

---

# 72. Forecast Guardrail

Before adding a forecast assumption, ask:

```text
Can the user understand why this number exists?
```

If not, the forecast is becoming too opaque.

---

# 73. Scenario Guardrail

Scenario modification code must work on:

```text
projection model
```

not:

```text
real repository write methods
```

where possible.

---

# 74. UI Guardrail

Every intelligence screen must distinguish:

```text
ACTUAL

PROJECTED

SCENARIO

AI
```

---

# 75. Error Handling Guardrail

Do not return:

```text
err.Error()
```

directly to client for unexpected errors.

Use application error mapping.

---

# 76. API Contract Guardrail

Do not silently change request/response fields during implementation.

If change required:

```text
update api-contract.md

update OpenAPI

update frontend types

update tests
```

---

# 77. Database Contract Guardrail

If implementation reveals a better schema:

```text
update database-design.md
```

Do not allow docs to remain knowingly incorrect.

---

# 78. Implementation Decisions Requiring ADR or Documentation Update

Document decisions if changing:

```text
authentication mechanism

balance source of truth

workspace model

forecast algorithm

scenario model

AI provider

queue system

transaction lifecycle
```

---

# 79. Recommended Development Order Within Each Day

A useful rhythm:

```text
1. Read milestone

2. Implement backend domain

3. Write backend tests

4. Add API

5. Integrate frontend

6. Add frontend tests

7. Update docs

8. Commit
```

---

# 80. Avoiding AI Coding Agent Drift

If coding agents are used, provide:

```text
specific milestone

source-of-truth docs

files allowed to change

acceptance criteria

required tests
```

Avoid prompts such as:

```text
Build Savio.
```

---

# 81. Coding Agent Task Format

Recommended:

```text
Objective

Source of Truth

Scope

Required Behavior

Out of Scope

Files / Modules

API Contract

Business Rules

Security Rules

Required Tests

Definition of Done
```

---

# 82. Example Coding Agent Task

```text
Objective:
Implement transaction creation and listing.

Source of Truth:
- docs/product/business-requirements.md
- docs/database/database-design.md
- docs/api/api-contract.md
- docs/engineering/security.md

Scope:
- POST /api/v1/transactions
- GET /api/v1/transactions

Rules:
- workspace scoped
- OWNER/MEMBER write
- VIEWER read only
- amount > 0
- income category must be INCOME
- expense category must be EXPENSE
- archived accounts rejected
- exact decimal arithmetic

Required Tests:
- happy path
- category mismatch
- cross-workspace
- viewer forbidden
- pagination
- filtering
```

---

# 83. Agent Verification Rule

Never accept agent output based only on:

```text
"implemented successfully"
```

Require:

```text
diff review

tests

build

runtime verification
```

---

# 84. Milestone Completion Verification

At each milestone run:

```bash
go test ./...
```

and relevant frontend checks.

At larger milestones:

```bash
go test -race ./...
npm run test
npm run build
```

---

# 85. P0 Demo Story

The implementation should ultimately support this coherent demo:

```text
1. Register

2. Default personal workspace created.

3. Add Bank Account.

4. Add monthly salary.

5. Add recurring rent.

6. Add several expenses.

7. Create Food budget.

8. Dashboard explains current cashflow.

9. Forecast shows expected balance trajectory.

10. Create "Buy Laptop" scenario.

11. Savio compares baseline vs scenario.

12. AI explains the main trade-off.

13. Ask Copilot:
    "Why is my spending higher this month?"

14. Copilot uses deterministic tools.

15. User reviews supporting facts.

16. User remains the decision-maker.
```

---

# 86. Secondary RBAC Demo Story

To demonstrate authorization:

```text
1. OWNER adds another existing Savio user.

2. Assign VIEWER.

3. VIEWER can open Dashboard.

4. VIEWER can inspect Transactions.

5. VIEWER cannot add expense.

6. Backend returns 403 if mutation attempted directly.

7. OWNER changes role to MEMBER.

8. MEMBER can now create transaction.
```

This demonstrates meaningful authorization beyond UI hiding.

---

# 87. Security Demo Story

Demonstrate verbally or through tests:

```text
Access token expires.

Three API calls return 401.

Frontend performs one refresh.

Refresh rotates token.

Requests retry successfully.

Old refresh token cannot be reused.
```

This is a strong technical-review scenario.

---

# 88. Concurrency Demo Story

Explain:

```text
Two concurrent financial writes hit the same account.

Ledger entries remain consistent.

Derived balance reflects both.

No mutable client-side balance overwrite exists.
```

---

# 89. AI Failure Demo Story

Disable AI provider.

Expected:

```text
Accounts work.

Transactions work.

Budgets work.

Analytics work.

Forecast works.

Scenario works.

AI features show graceful degraded state.
```

This proves architectural boundaries.

---

# 90. Submission Quality Targets

The completed repository should feel:

```text
intentional

reviewable

coherent

secure

tested

documented
```

not:

```text
feature-stuffed
```

---

# 91. Final Repository Target

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
│   ├── internal/
│   │   ├── platform/
│   │   ├── auth/
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

# 92. Final Pre-Coding Checklist

Before starting implementation:

```text
[ ] Product thesis locked

[ ] P0 scope locked

[ ] Business rules documented

[ ] User flows documented

[ ] Workspace/RBAC model locked

[ ] Balance source-of-truth locked

[ ] Transaction lifecycle locked

[ ] Recurring semantics locked

[ ] Forecast algorithm baseline locked

[ ] Scenario semantics locked

[ ] AI authority boundary locked

[ ] Database design reviewed

[ ] API contract reviewed

[ ] Security architecture reviewed

[ ] Testing strategy reviewed

[ ] Design system reviewed

[ ] Implementation milestones reviewed
```

---

# 93. Final Scope Lock

The implementation must not expand casually once coding starts.

Any new feature must answer:

```text
Does it directly improve assessment quality?

Does it strengthen Savio's product thesis?

Is P0 already stable?

Can it be tested and documented properly?
```

If not:

```text
defer it.
```

---

# 94. Final Implementation Definition of Success

Savio is implementation-ready when the development team can answer all of these without inventing requirements during coding:

```text
What are we building?

Who is it for?

What is P0?

What is authoritative financial data?

How is money represented?

How is balance calculated?

How do transfers behave?

How do recurring transactions behave?

How are budgets calculated?

How does forecasting work?

How does a scenario work?

What may AI do?

What may AI never do?

How does authentication work?

How does refresh rotation work?

How does CSRF work?

How are roles enforced?

How are resources scoped?

What happens under concurrency?

What is tested?

What is deployed?
```

---

# 95. Final Implementation Principle

Implementation should follow the product's authority hierarchy:

```text
SECURITY
    ↓
AUTHORIZATION
    ↓
AUTHORITATIVE FINANCIAL DATA
    ↓
BUSINESS RULES
    ↓
DETERMINISTIC FINANCE ENGINE
    ↓
FORECAST
    ↓
SCENARIO
    ↓
AI INTERPRETATION
    ↓
FRONTEND PRESENTATION
    ↓
USER DECISION
```

The development process should preserve that hierarchy from the first migration to the final UI.

The final Savio rule remains:

> **Finance Engine calculates. AI interprets. User decides.**
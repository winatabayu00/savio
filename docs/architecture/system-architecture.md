# Savio — System Architecture

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Business Requirements](../product/business-requirements.md) — domain boundaries this architecture implements.
- [Database Design](../database/database-design.md) — persistence layer mapped from this architecture.
- [API Contract](../api/api-contract.md) — endpoints exposed by the application layers.
- [Frontend Architecture](frontend-architecture.md) — frontend counterpart of these layers.
- [AI Architecture](ai-architecture.md) — AI subsystem placement within the modular monolith.
- [Security Architecture](../engineering/security.md) — security boundaries across the system.

## 1. Document Purpose

This document defines the overall system architecture for Savio.

The purpose of this document is to translate the product foundation, business requirements, user flows, database design, API contract, and AI architecture into one coherent technical architecture.

This document defines:

- application architecture,
- backend structure,
- frontend integration,
- finance engine boundaries,
- AI integration,
- PostgreSQL usage,
- Redis usage,
- background workers,
- object storage,
- authentication,
- authorization,
- CSRF,
- validation,
- transaction boundaries,
- concurrency,
- error handling,
- observability,
- performance,
- deployment,
- testing,
- reliability,
- scaling,
- and engineering trade-offs.

The core Savio principle remains:

> **Finance Engine calculates. AI interprets. User decides.**

---

# 2. System Overview

Savio is designed as:

> **A modular monolith with a dedicated background worker and an external AI provider integration.**

Primary components:

```text
React Frontend
     ↓
Go REST API
     ↓
Modular Application Services
     ↓
Finance Engine
     ↓
PostgreSQL

Additional infrastructure:

Redis
Go Worker
MinIO
AI Provider
```

---

# 3. High-Level Architecture

```text
                               USER
                                │
                                ▼
                     ┌────────────────────┐
                     │   React Frontend   │
                     │ Vite + TypeScript  │
                     └──────────┬─────────┘
                                │
                          HTTPS / REST
                                │
                                ▼
                     ┌────────────────────┐
                     │      Go API        │
                     │    Gin + GORM      │
                     └──────────┬─────────┘
                                │
        ┌───────────────────────┼────────────────────────┐
        │                       │                        │
        ▼                       ▼                        ▼
┌────────────────┐     ┌────────────────┐      ┌─────────────────┐
│ Domain Modules │     │ Finance Engine │      │ AI Orchestrator │
└───────┬────────┘     └───────┬────────┘      └────────┬────────┘
        │                      │                        │
        └──────────────┬───────┘                        │
                       │                                │
                       ▼                                ▼
               ┌────────────────┐              ┌─────────────────┐
               │   PostgreSQL   │              │   AI Provider   │
               └────────────────┘              └─────────────────┘

                       │
                       │ async jobs
                       ▼
                 ┌─────────────┐
                 │    Redis    │
                 └──────┬──────┘
                        │
                        ▼
                 ┌─────────────┐
                 │  Go Worker  │
                 └──────┬──────┘
                        │
              ┌─────────┼─────────┐
              ▼         ▼         ▼
         PostgreSQL   MinIO   AI Provider
```

---

# 4. Architecture Style

Savio uses a:

```text
MODULAR MONOLITH
```

rather than:

```text
MICROSERVICES
```

for the initial implementation.

---

# 5. Why Modular Monolith

The project requires:

- transactional financial consistency,
- a relatively small engineering footprint,
- simple local setup,
- clear reviewability,
- reliable testing,
- and low deployment complexity.

A modular monolith provides:

```text
clear module boundaries
+
single deployment unit
+
shared transactional database
+
simple debugging
+
lower infrastructure complexity
```

without forcing distributed-system complexity prematurely.

---

# 6. Why Not Microservices

Microservices would introduce:

```text
network boundaries
service discovery
distributed tracing
distributed transactions
message contracts
deployment coordination
failure propagation
more infrastructure
```

without a clear need for Savio's initial scale.

For a take-home scope, this would increase complexity without improving the core product.

---

# 7. Application Boundaries

The system is split conceptually into:

```text
Frontend

Backend API

Background Worker

Database

Queue / Cache

Object Storage

External AI Provider
```

---

# 8. Repository-Level Architecture

Recommended repository:

```text
savio/
├── README.md
├── AGENTS.md
├── DESIGN.md
│
├── backend/
│   ├── cmd/
│   │   ├── api/
│   │   └── worker/
│   │
│   ├── internal/
│   ├── pkg/
│   ├── migrations/
│   ├── seeds/
│   └── tests/
│
├── frontend/
│   ├── src/
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
├── .github/
├── docker-compose.yml
├── Makefile
└── .env.example
```

Detailed repository rules will be defined later in the implementation plan and `AGENTS.md`.

---

# 9. Backend Architecture

Recommended backend flow:

```text
HTTP Request
    ↓
Middleware
    ↓
Handler
    ↓
Service
    ↓
Domain / Finance Engine
    ↓
Repository
    ↓
GORM
    ↓
PostgreSQL
```

Cross-cutting components:

```text
Auth
Validation
Authorization
Logging
Rate Limit
Transactions
Errors
Audit
Queue
AI
```

---

# 10. Backend Module Structure

Recommended modules:

```text
internal/
├── auth/
├── sessions/
├── users/
├── settings/
│
├── accounts/
├── categories/
├── transactions/
├── transfers/
├── recurring/
│
├── budgets/
├── goals/
│
├── analytics/
├── forecast/
├── scenarios/
│
├── ai/
├── notifications/
├── audit/
│
└── platform/
```

---

# 11. Platform Package

Shared technical infrastructure belongs in a platform/shared area.

Example:

```text
internal/platform/
├── config/
├── database/
├── logger/
├── middleware/
├── response/
├── validator/
├── errors/
├── security/
├── queue/
├── storage/
├── pagination/
├── clock/
└── telemetry/
```

These packages should not own Savio business rules.

---

# 12. Handler Layer

Handlers are responsible for:

```text
HTTP concerns
```

including:

- reading path parameters,
- reading query parameters,
- decoding request bodies,
- triggering validation,
- extracting authenticated user context,
- calling application services,
- mapping service errors to HTTP responses,
- and serializing output.

Handlers should be intentionally thin.

---

# 13. Handler Anti-Pattern

Avoid:

```go
func CreateTransaction(c *gin.Context) {
    // parse request
    // query account
    // calculate balance
    // validate category
    // update account
    // create audit
    // mark forecast stale
    // enqueue job
    // ...
}
```

This mixes transport and business logic.

Preferred:

```text
Handler
    ↓
TransactionService.Create(...)
```

---

# 14. Service Layer

Services own:

- business rules,
- orchestration,
- transaction boundaries,
- authorization checks,
- deterministic calculations,
- state transitions,
- and collaboration between repositories/modules.

Example:

```text
TransactionService

TransferService

RecurringService

BudgetService

GoalService

ForecastService

ScenarioService
```

---

# 15. Repository Layer

Repositories are responsible for:

```text
database persistence
query construction
row locking
aggregate queries
```

Repositories should not own high-level business policy.

Example:

```text
AccountRepository

TransactionRepository

BudgetRepository

ForecastRepository
```

---

# 16. Domain Logic

Business calculations should live in dedicated domain or finance-engine components where possible.

Example:

```text
CalculateSavingsRate()

CalculateBudgetUtilization()

CalculateGoalProgress()

CalculateForecast()

CalculateScenario()
```

These should ideally be testable without HTTP or AI dependencies.

---

# 17. Finance Engine

The Finance Engine is the deterministic intelligence core of Savio.

It owns:

```text
Account balance rules

Cashflow calculation

Savings rate

Budget utilization

Budget projection

Goal progress

Goal feasibility

Cash runway

Forecasting

Scenario simulation

Financial health calculations
```

---

# 18. Finance Engine Architecture

```text
Financial Records
      ↓
Finance Engine
      │
      ├── Balance Engine
      ├── Analytics Engine
      ├── Budget Engine
      ├── Goal Engine
      ├── Forecast Engine
      ├── Scenario Engine
      └── Health Engine
      ↓
Structured Financial Results
```

---

# 19. Finance Engine Dependency Rule

The Finance Engine may depend on:

```text
domain types
deterministic configuration
clock
decimal utilities
```

It should not depend on:

```text
AI provider
HTTP handler
React frontend
external chat model
```

---

# 20. AI Dependency Direction

Correct:

```text
AI
 ↓
Finance Engine
```

Incorrect:

```text
Finance Engine
 ↓
AI
```

This guarantees that finance functionality remains available when AI is unavailable.

---

# 21. AI Architecture Integration

AI sits above deterministic finance services.

```text
User
 ↓
AI Request
 ↓
AI Orchestrator
 ↓
Finance Tools
 ↓
Finance Engine
 ↓
Structured Facts
 ↓
Context Builder
 ↓
AI Provider
 ↓
Validated Interpretation
```

---

# 22. Background Worker

Savio uses a separate worker process:

```text
backend/cmd/worker
```

The worker runs from the same codebase and can reuse:

- services,
- repositories,
- configuration,
- AI orchestration,
- queue infrastructure.

---

# 23. API and Worker Deployment

Conceptually:

```text
same codebase
different entry points
```

Example:

```text
backend/cmd/api/main.go

backend/cmd/worker/main.go
```

This avoids creating a separate service repository while still separating synchronous and asynchronous execution.

---

# 24. Worker Responsibilities

Worker tasks may include:

```text
Recurring transaction posting

Recurring occurrence processing

Upcoming bill reminders

Budget risk detection

Forecast risk evaluation

AI insight generation

Notification generation

Expired auth session cleanup

Future import processing

Future receipt extraction
```

---

# 25. Worker Non-Responsibilities

Worker should not become a second business implementation.

It should call the same services used by the API.

Example:

Correct:

```text
Worker
 ↓
RecurringService.ProcessDueOccurrence()
```

Avoid:

```text
Worker contains duplicate balance logic
```

---

# 26. Redis

Redis may be used for:

```text
background queue
rate limiting
short-lived cache
distributed coordination if needed
```

Redis is not the authoritative financial database.

---

# 27. Redis Source-of-Truth Rule

Never store authoritative only-copy financial data in Redis.

The following remain in PostgreSQL:

```text
accounts

transactions

transfers

budgets

goals

recurring rules

forecast snapshots

scenario snapshots
```

---

# 28. Queue Implementation

A Redis-backed Go job framework may be used.

Possible choice:

```text
Asynq
```

or another well-understood Redis-backed queue.

The implementation should support:

```text
retry
idempotency
scheduled execution
failed job visibility
```

The exact library is an implementation choice.

---

# 29. Queue Job Design

Jobs should contain minimal identifiers.

Better:

```json
{
  "user_id": "uuid",
  "signal_id": "uuid"
}
```

rather than:

```json
{
  "full_financial_dataset": "..."
}
```

Worker reloads authoritative data when needed.

---

# 30. Queue Idempotency

Background jobs must be safe to retry.

Examples:

```text
Recurring posting

AI insight generation

Notification generation
```

All should have deterministic deduplication keys where relevant.

---

# 31. Queue Failure Principle

Queue failure must not corrupt core financial state.

Example:

```text
Expense successfully committed
AI queue temporarily unavailable
```

Expected:

```text
expense remains valid
```

The user should not receive a financial failure after the transaction has already committed.

---

# 32. Post-Commit Side Effects

Typical flow:

```text
BEGIN DB TRANSACTION
    ↓
Financial Write
    ↓
Audit / freshness update
    ↓
COMMIT
    ↓
Attempt async enqueue
```

If enqueue fails:

```text
log failure
```

and potentially recover later.

---

# 33. Transactional Outbox

Future reliability improvement:

```text
financial transaction
+
outbox event
```

written in one PostgreSQL transaction.

Worker then publishes/handles events.

This removes the commit-to-enqueue failure window.

For the initial implementation:

```text
post-commit queue dispatch
```

is acceptable if documented as a trade-off.

---

# 34. PostgreSQL

PostgreSQL is the system of record for Savio.

It stores:

```text
identity

sessions

financial records

configuration

forecast snapshots

scenario snapshots

AI insight metadata

notifications

audit logs
```

---

# 35. GORM

GORM is used as the backend ORM.

GORM responsibilities:

```text
model mapping

queries

transactions

row locking

preloading where appropriate
```

GORM AutoMigrate should not be the production migration source of truth.

---

# 36. Database Migration

Schema migrations use explicit SQL.

Recommended:

```text
golang-migrate
```

Commands should eventually provide:

```text
make migrate-up

make migrate-down
```

The schema must be reproducible from an empty database.

---

# 37. MinIO / Object Storage

MinIO is used for file storage when file capabilities are implemented.

Potential use cases:

```text
receipt images

CSV imports

future bank statements

report exports
```

Binary objects:

```text
MinIO
```

Metadata:

```text
PostgreSQL
```

---

# 38. File Storage Principle

Do not store persistent user uploads in:

```text
local backend filesystem
```

because:

- containers are ephemeral,
- horizontal scaling becomes difficult,
- file ownership is less explicit.

Use object storage.

---

# 39. Authentication Architecture

Authentication uses:

```text
short-lived access credential
+
rotating refresh token
+
server-side refresh sessions
+
secure cookies
```

Recommended model:

```text
Access:
JWT

Refresh:
opaque random token
```

---

# 40. Why Opaque Refresh Tokens

Opaque refresh tokens are recommended because:

```text
server-side revocation is natural

rotation is straightforward

no refresh-token JWT parsing needed

session lifecycle remains explicit
```

Only a secure hash is stored.

---

# 41. Auth Request Flow

```text
Login
  ↓
Validate credentials
  ↓
Create auth session
  ↓
Generate access JWT
  ↓
Generate random refresh token
  ↓
Hash refresh token
  ↓
Store session
  ↓
Set cookies
```

---

# 42. Access Token

Access token contains minimum claims.

Example:

```text
sub
session_id
iat
exp
```

Avoid embedding large authorization state if not required.

---

# 43. Refresh Flow

```text
Refresh Cookie
    ↓
Hash / verify
    ↓
Load Session
    ↓
Not revoked?
    ↓
Not expired?
    ↓
Rotate token
    ↓
Update stored hash
    ↓
Issue new access token
    ↓
Issue new refresh token
```

---

# 44. Refresh Rotation Race

Concurrent refresh requests can create problems.

Example:

```text
Request A
Request B
```

both using the same refresh token.

The backend must define atomic rotation semantics.

Possible implementation:

```text
transaction / row lock auth_session
```

Only one valid rotation should win.

Frontend single-flight refresh further minimizes this condition.

---

# 45. Cookie Security

Production cookies:

```text
HttpOnly
Secure
SameSite=Lax
explicit Path
finite Max-Age / Expires
```

Local development may set:

```text
Secure=false
```

only when HTTPS is unavailable.

---

# 46. CSRF Architecture

Because authentication uses cookies, CSRF protection is mandatory.

Recommended:

```text
signed double-submit token
```

or equivalent session-bound CSRF token.

---

# 47. CSRF Flow

```text
Frontend obtains CSRF token
      ↓
Readable CSRF cookie / bootstrap endpoint
      ↓
Frontend sends:
X-CSRF-Token
      ↓
Backend verifies
      ↓
State-changing request continues
```

---

# 48. CSRF Protected Methods

Protect:

```text
POST
PUT
PATCH
DELETE
```

Safe reads:

```text
GET
HEAD
OPTIONS
```

do not normally require CSRF validation.

---

# 49. Login CSRF

Login itself changes authentication state.

The security design should also protect login against cross-site request abuse.

A pre-auth CSRF bootstrap endpoint can provide a valid CSRF token before login.

---

# 50. Refresh CSRF

Refresh also uses automatically attached cookies and should be protected appropriately.

The chosen token path and CSRF strategy must support refresh without weakening protection.

---

# 51. Authorization

Savio's primary MVP is user-owned personal finance.

Authorization therefore depends heavily on:

```text
authentication
+
resource ownership
```

rather than broad internal RBAC.

---

# 52. Resource Authorization

Every user-owned resource must satisfy:

```text
resource.user_id
=
authenticated_user.id
```

This includes:

```text
accounts
transactions
transfers
recurring rules
budgets
goals
forecasts
scenarios
AI insights
notifications
sessions
```

---

# 53. RBAC Extension

The take-home requirement expects multiple authorization levels.

Savio can support system-level roles such as:

```text
USER

ADMIN
```

Potential future:

```text
HOUSEHOLD_OWNER
HOUSEHOLD_MEMBER
HOUSEHOLD_VIEWER
```

The MVP should still enforce meaningful backend authorization rather than role checks existing only in UI.

---

# 54. Admin Role Scope

Admin must not automatically receive unrestricted access to private financial records unless explicitly necessary.

A safe initial admin capability might include:

```text
system health

user account status

AI provider diagnostics

job diagnostics
```

without browsing arbitrary transaction contents.

This is more privacy-conscious.

---

# 55. RBAC + Ownership

Conceptually:

```text
Permission
+
Resource Ownership
```

Example:

```text
USER has transaction.update

but

transaction.user_id must equal USER.id
```

Permission alone is insufficient.

---

# 56. Authorization Middleware

Use middleware for coarse requirements:

```text
authenticated?
required role?
required permission?
```

Use service/resource policy for:

```text
does this resource belong to user?
```

---

# 57. Validation Architecture

Validation occurs at multiple levels.

```text
Frontend
    ↓
HTTP DTO Validation
    ↓
Service Business Validation
    ↓
Database Constraints
```

Each has a different purpose.

---

# 58. Frontend Validation

Frontend validates for UX:

```text
required fields

format

min/max

known enum
```

But can be bypassed.

---

# 59. DTO Validation

Backend validates request structure.

Examples:

```text
UUID

email

amount

date

enum

string length

pagination

sort field
```

---

# 60. Business Validation

Services validate contextual rules.

Examples:

```text
expense category matches EXPENSE

account belongs to user

account is ACTIVE

transfer accounts differ

scenario modification fields match type

budget does not conflict
```

---

# 61. Database Validation

PostgreSQL enforces simple invariants:

```text
foreign key

unique constraints

amount > 0

date range

source account != destination
```

---

# 62. Error Architecture

Savio should use centralized application errors.

Concept:

```go
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
    Details    any
    Cause      error
}
```

Internal `Cause` should not be serialized to users.

---

# 63. Error Categories

```text
ValidationError

AuthenticationError

AuthorizationError

NotFoundError

ConflictError

BusinessRuleError

RateLimitError

ExternalDependencyError

InternalError
```

---

# 64. Error Flow

```text
Repository / Service Error
      ↓
App Error Mapping
      ↓
Handler
      ↓
Standard API Error Response
```

Unexpected error:

```text
log full internal detail
```

Public response:

```text
sanitized 500
```

---

# 65. Financial Transaction Boundaries

Operations affecting balances must be atomic.

Critical examples:

```text
Create income

Create expense

Edit transaction

Void transaction

Create transfer

Void transfer

Post recurring occurrence

Reconcile account
```

---

# 66. Create Expense Transaction

Conceptual flow:

```text
BEGIN
    ↓
Load + lock account
    ↓
Validate user ownership
    ↓
Validate category
    ↓
Insert transaction
    ↓
Update balance atomically
    ↓
Mark forecasts stale
    ↓
Mark scenarios stale
    ↓
Create audit
COMMIT
```

---

# 67. Transfer Transaction

```text
BEGIN
    ↓
Lock source account
    ↓
Lock destination account
    ↓
Validate ownership
    ↓
Validate active status
    ↓
Update source
    ↓
Update destination
    ↓
Create transfer
    ↓
Invalidate derived snapshots
    ↓
Audit
COMMIT
```

---

# 68. Account Lock Ordering

Concurrent transfers may deadlock if rows are locked inconsistently.

Example:

```text
Transfer A:
Account 1 → Account 2

Transfer B:
Account 2 → Account 1
```

To reduce deadlock risk:

```text
lock account rows in deterministic ID order
```

regardless of transfer direction.

---

# 69. Balance Update Strategy

Use:

```text
row lock
```

or:

```text
atomic SQL arithmetic
```

Avoid naive:

```text
read current_balance
↓
modify in Go
↓
overwrite
```

without concurrency protection.

---

# 70. Optimistic Locking

Important mutable resources carry:

```text
version
```

Client reads:

```text
version = 5
```

Update submits:

```text
version = 5
```

SQL:

```text
WHERE version = 5
```

on success:

```text
version = 6
```

---

# 71. Optimistic Lock vs Pessimistic Lock

Use optimistic locking for:

```text
user form editing
budgets
goals
recurring configuration
scenarios
```

Use row locks / atomic database writes for:

```text
balance-affecting financial operations
```

because those require stronger transactional serialization.

---

# 72. Forecast Architecture

The Forecast Engine consumes authoritative data and creates projected financial events.

```text
Accounts
Recurring Rules
Scheduled Events
Historical Spending
Assumptions
      ↓
Forecast Input Builder
      ↓
Event Generator
      ↓
Event Ordering
      ↓
Balance Projection
      ↓
Risk Metrics
      ↓
Forecast Result
```

---

# 73. Forecast Event Types

```text
KNOWN

SCHEDULED

ESTIMATED

ASSUMED
```

These classifications must remain visible in the result.

---

# 74. Forecast Calculation Boundary

The forecast engine performs all financial arithmetic.

AI receives forecast output only after the forecast completes.

---

# 75. Forecast Freshness

Financial writes invalidate previous forecasts.

Simple strategy:

```text
mark existing forecast snapshots STALE
```

A background or on-demand recalculation can later create a fresh snapshot.

---

# 76. Forecast Live vs Snapshot

Savio may support both:

```text
live calculation
```

and:

```text
persisted snapshot
```

Live:

```text
useful for dashboard preview
```

Snapshot:

```text
useful for reproducibility
scenario linkage
historical review
```

---

# 77. Scenario Architecture

Scenario simulation operates on a copy/model of the financial baseline.

It does not modify real financial data.

```text
Current Financial State
      ↓
Baseline Forecast
      ↓
Scenario Modification Set
      ↓
Apply Modifications
      ↓
Recalculate
      ↓
Compare Baseline vs Scenario
```

---

# 78. Scenario Isolation

The scenario layer must never write:

```text
real transaction

real account balance

real recurring rule

real budget
```

during calculation.

Only scenario entities/snapshots may be persisted.

---

# 79. Scenario Engine Design

Potential internal types:

```text
ScenarioInput

Baseline

ScenarioModification

ProjectionEvent

ScenarioResult

ScenarioDifference
```

These can be pure domain structs to make testing easy.

---

# 80. Analytics Architecture

Analytics should use PostgreSQL aggregation where appropriate.

Example:

```text
sum income

sum expense

category totals

period comparisons
```

Do not load every transaction into Go for simple group-by queries.

---

# 81. Analytics Service

Analytics service may combine:

```text
repository aggregate results
+
finance formulas
```

Example:

```text
PostgreSQL:
SUM income and expense

Go Finance Engine:
calculate savings rate
```

---

# 82. Dashboard Architecture

Dashboard is a composite read model.

Potential dependencies:

```text
Account Summary

Cashflow Analytics

Budgets

Goals

Recurring Events

Forecast Preview

AI Insights
```

---

# 83. Composite Endpoint

Recommended:

```text
GET /api/v1/dashboard
```

Backend assembles coherent screen data.

This avoids excessive frontend waterfall requests.

---

# 84. Dashboard Performance

Avoid:

```text
10 sequential repository queries
```

where independent reads can run safely in parallel.

However:

```text
parallelism should remain bounded
```

and complexity should only be introduced where useful.

---

# 85. Frontend Architecture Overview

Frontend stack:

```text
React
TypeScript
Vite
Tailwind CSS
Axios
React Router
TanStack Query
React Hook Form
Zod
```

---

# 86. Frontend Responsibility Boundary

Frontend owns:

```text
presentation

user interaction

client validation

navigation

loading states

error states

server-state caching

formatting
```

Frontend does not own authoritative financial logic.

---

# 87. Server State

Use:

```text
TanStack Query
```

for server state.

Avoid duplicating server entities into large custom global stores unless genuinely needed.

---

# 88. Form State

Use:

```text
React Hook Form
+
Zod
```

for:

```text
form state

client-side validation

field error rendering
```

Backend validation remains authoritative.

---

# 89. Axios Architecture

Central Axios client should configure:

```text
base URL

credentials: include

CSRF header

request ID handling if needed

interceptors
```

---

# 90. 401 Interceptor

Required behavior:

```text
401
 ↓
single-flight refresh
 ↓
queue concurrent failed requests
 ↓
refresh succeeds
 ↓
retry once
```

Failed refresh:

```text
logout
```

---

# 91. 403 Interceptor

Do not logout automatically.

Show:

```text
permission / CSRF error
```

depending on error code.

---

# 92. 422 Interceptor

Pass structured field validation details to forms.

---

# 93. 429 Interceptor

Display rate-limit feedback.

Do not blindly retry AI or login requests.

---

# 94. 500 Interceptor

Show safe generic error.

Potentially include:

```text
request ID
```

for support/debugging.

---

# 95. Frontend Query Keys

Recommended domain-centered query keys.

Example:

```text
["accounts"]

["account", id]

["transactions", filters]

["budgets", period]

["forecast", horizon]

["scenario", id]
```

---

# 96. Query Invalidation

Example after expense creation:

```text
invalidate accounts

invalidate transactions

invalidate dashboard

invalidate analytics

invalidate budgets

invalidate latest forecast status

invalidate scenarios where relevant
```

Do not rely on stale client calculations.

---

# 97. UI State Architecture

Every significant screen should support:

```text
Loading

Empty

Success

Error

Disabled / Pending

Responsive
```

---

# 98. Security Middleware Order

Recommended request middleware order:

```text
Request ID
    ↓
Recovery
    ↓
Security Headers
    ↓
CORS
    ↓
Structured Logging
    ↓
Rate Limit
    ↓
Authentication
    ↓
CSRF
    ↓
Authorization
    ↓
Handler
```

Exact order may vary slightly, but security dependencies should remain intentional.

---

# 99. Authentication Middleware

Responsibilities:

```text
read access cookie

validate JWT

load session context if needed

set authenticated user context
```

Invalid token:

```text
401
```

---

# 100. CSRF Middleware

Runs for state-changing authenticated endpoints.

Responsibilities:

```text
read token cookie
read token header
validate signature/session binding
constant-time compare if applicable
```

---

# 101. Rate Limiting Architecture

Redis-backed rate limiting may be used.

Scopes:

```text
IP

user

endpoint group
```

Examples:

```text
login
register
refresh

AI Copilot

AI categorization
```

---

# 102. Brute Force Protection

Login protection may combine:

```text
IP-based rate limit

account/email-based rate limit

increasing cooldown
```

Do not expose whether an email exists through rate-limit messaging.

---

# 103. Password Hashing

Use a modern password hashing algorithm.

Recommended options:

```text
Argon2id
```

or:

```text
bcrypt
```

with appropriately configured cost.

---

# 104. Secrets

Secrets come from environment variables or secrets management.

Examples:

```text
DATABASE_URL

REDIS_URL

JWT_SECRET

CSRF_SECRET

AI_API_KEY

MINIO_ACCESS_KEY

MINIO_SECRET_KEY
```

Never commit production secrets.

---

# 105. Configuration

Configuration should be centralized.

Concept:

```go
type Config struct {
    App      AppConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Auth     AuthConfig
    AI       AIConfig
    Storage  StorageConfig
}
```

Validate configuration at startup.

Fail fast on missing critical configuration.

---

# 106. Environment Strategy

Possible environments:

```text
development

test

production
```

Behavior differences should be explicit.

Examples:

```text
Secure cookies
logging level
AI mock provider
CORS origins
```

---

# 107. Development Environment

Docker Compose should provide:

```text
PostgreSQL

Redis

MinIO
```

Frontend and backend may run:

```text
inside Docker
```

or:

```text
locally
```

depending on developer preference.

Final README should document one clear path.

---

# 108. Docker Compose

Recommended services:

```text
frontend

api

worker

postgres

redis

minio
```

Optional:

```text
minio-setup
```

for bucket initialization.

---

# 109. Docker Network

All services share an internal Docker network.

Example service resolution:

```text
postgres:5432

redis:6379

minio:9000
```

---

# 110. Backend Container

Backend image should use:

```text
multi-stage build
```

Example:

```text
builder
→ compile Go binary

runtime
→ small final image
```

---

# 111. Frontend Container

Frontend can use:

```text
Node build stage
↓
static assets
↓
Nginx / lightweight static server
```

or be served independently.

---

# 112. Database Startup

Application startup should not blindly perform production schema migration unless intentionally designed.

Preferred developer flow:

```text
infrastructure starts
↓
migration command
↓
application starts
```

This keeps schema control explicit.

---

# 113. Health Architecture

Endpoints:

```text
/health

/ready
```

---

# 114. Liveness

`/health` checks:

```text
process is alive
```

It should not fail simply because AI provider is down.

---

# 115. Readiness

`/ready` checks critical dependencies.

PostgreSQL:

```text
required
```

Redis:

```text
may be required or degraded depending on how essential worker features are
```

AI provider:

```text
degraded only
```

because finance features work without it.

---

# 116. Dependency Health Model

Example:

```json
{
  "status": "ready",
  "dependencies": {
    "postgres": "up",
    "redis": "up",
    "minio": "up",
    "ai": "degraded"
  }
}
```

---

# 117. Observability

Minimum observability:

```text
structured logs

request ID

health

readiness

worker logs

AI request metrics/logs
```

---

# 118. Structured Logging

Use structured JSON logs in production.

Fields may include:

```text
timestamp

level

request_id

user_id

method

path

status

duration_ms

error_code
```

---

# 119. Sensitive Logging Rules

Do not log:

```text
password

access token

refresh token

CSRF secret

database credentials

AI API key

full financial request bodies by default
```

---

# 120. Request ID

Every incoming request gets:

```text
request_id
```

If client supplies a valid request ID format, it may be reused or replaced depending on policy.

All logs and AI calls should correlate with the same request.

---

# 121. Worker Job Correlation

Worker jobs should include:

```text
job_id

originating request_id where available

user_id

job type
```

This helps trace asynchronous behavior.

---

# 122. AI Observability

Track:

```text
provider

model

feature

latency

status

token usage

error code
```

Avoid storing full sensitive context.

---

# 123. Metrics

Optional future metrics:

```text
http_requests_total

http_request_duration

db_query_duration

queue_jobs_total

queue_failures_total

ai_requests_total

ai_latency

ai_errors
```

Prometheus/OpenTelemetry can be P2.

---

# 124. Tracing

Full distributed tracing is not required initially because Savio is a modular monolith.

If added later, useful spans include:

```text
HTTP request

database

Redis

AI provider

MinIO

worker job
```

---

# 125. Performance Strategy

Performance optimizations should follow observed bottlenecks.

Initial priorities:

```text
database indexes

bounded pagination

aggregate queries

N+1 prevention

connection pooling

context cancellation

reasonable API timeouts
```

---

# 126. N+1 Prevention

Example bad behavior:

```text
20 transactions
↓
20 account queries
↓
20 category queries
```

Preferred:

```text
join / preload / bulk lookup
```

---

# 127. Database Connection Pool

Configure:

```text
max open connections

max idle connections

connection lifetime
```

based on deployment environment.

Do not accept default behavior blindly for production.

---

# 128. Query Context

Database operations should use:

```text
request-scoped context
```

to support cancellation and deadlines.

---

# 129. API Timeouts

Application/server timeouts should include:

```text
read timeout

write timeout

idle timeout

external AI timeout
```

AI call timeout should be much shorter than an unlimited HTTP request.

---

# 130. Pagination

Use bounded offset pagination for MVP.

Example:

```text
page

limit
```

Maximum:

```text
100
```

Future large append-only datasets may use cursor pagination.

---

# 131. Cache Strategy

Do not introduce caching before needed.

Potential future candidates:

```text
dashboard aggregate

system categories

AI merchant categorization
```

Cache must never become authoritative financial state.

---

# 132. Cache Invalidation

If dashboard is cached:

```text
transaction create
transaction update
transfer
budget update
goal update
```

must invalidate relevant cache keys.

This complexity is why caching is P2.

---

# 133. Search

Initial transaction search may use:

```text
ILIKE
```

over selected fields.

If data volume grows:

```text
pg_trgm
```

may be added.

Elasticsearch is unnecessary for MVP.

---

# 134. File Upload Architecture

Future receipt upload:

```text
Frontend
   ↓
Backend
   ↓
Validate file
   ↓
MinIO
   ↓
PostgreSQL metadata
```

---

# 135. File Validation

Validate:

```text
authentication

ownership

maximum size

MIME type

extension

content where practical

object key
```

Never trust original file name as storage path.

---

# 136. MinIO Object Keys

Example safe format:

```text
users/{user_id}/receipts/{attachment_id}.jpg
```

Avoid:

```text
../../filename
```

or direct unsanitized user path.

---

# 137. Object Access

Do not make financial receipts public by default.

Access should go through:

```text
authorized backend
```

or short-lived signed URLs generated after authorization.

---

# 138. Background Receipt Processing

Future:

```text
Upload
 ↓
Persist metadata
 ↓
Queue extraction job
 ↓
Worker
 ↓
AI/OCR
 ↓
Draft transaction
 ↓
User confirmation
```

No direct authoritative transaction creation.

---

# 139. Security Headers

Recommended:

```text
X-Content-Type-Options: nosniff

Referrer-Policy

Content-Security-Policy

Permissions-Policy

Strict-Transport-Security
```

Exact CSP depends on frontend deployment.

---

# 140. CORS

Credentialed CORS must use explicit origins.

Correct:

```text
https://savio.example.com
```

Avoid:

```text
*
```

with cookies.

---

# 141. Same-Origin Deployment

If possible, deployment under the same site simplifies:

```text
cookies

CORS

CSRF
```

Example:

```text
https://savio.example.com

https://savio.example.com/api/v1
```

Reverse proxy can route API traffic internally.

---

# 142. Alternative Split-Origin Deployment

Example:

```text
app.savio.example.com

api.savio.example.com
```

Still same-site depending on cookie domain configuration.

Requires careful CORS and cookie settings.

---

# 143. Financial Data Privacy

Savio stores sensitive personal financial data.

Design principles:

```text
least exposure

ownership scoping

minimal logging

minimal AI context

secure cookies

server-side provider keys

no public financial resources
```

---

# 144. Data Encryption

Transport:

```text
HTTPS
```

At rest:

```text
PostgreSQL / disk encryption provided by deployment infrastructure
```

Application-level encryption may be considered later for highly sensitive fields, but should not be added without clear key-management strategy.

---

# 145. Database User

Runtime API should use a dedicated database role.

Avoid:

```text
postgres superuser
```

for normal application execution.

Migration process may use a separate role if desired.

---

# 146. Idempotency

Financial creation commands can support:

```text
Idempotency-Key
```

to protect against duplicate retries.

Candidate:

```text
transaction create

transfer create
```

---

# 147. Idempotency Persistence

If implemented, use a table or durable mechanism.

Conceptual:

```text
idempotency_keys

user_id
key
route
request_hash
response_status
response_body
expires_at
```

This may be P1 if scope is tight.

Recurring worker idempotency remains required regardless.

---

# 148. Duplicate Submission Protection

Frontend:

```text
disable submit while pending
```

Backend:

```text
idempotency / unique rules where applicable
```

Frontend protection alone is insufficient.

---

# 149. Background Job Reliability

Each job should define:

```text
retry count

retry delay

idempotency key

failure logging
```

Example:

```text
AI insight generation
→ retry transient AI errors
```

---

# 150. Job Retry Classification

Retry:

```text
network failure

AI 502

Redis temporary error
```

Do not endlessly retry:

```text
invalid financial data

malformed job payload

deleted user
```

---

# 151. Failed Jobs

Failed jobs should remain visible through:

```text
queue dashboard/logging
```

or structured application logs.

The project should be able to explain:

```text
what happens when a worker job fails?
```

---

# 152. Recurring Scheduler

Worker should periodically query:

```text
ACTIVE recurring rules
where next_occurrence_date <= today
```

with proper indexing.

---

# 153. Scheduler Concurrency

Multiple worker instances must not post the same occurrence twice.

Use:

```text
unique recurring occurrence constraint
+
transaction
```

as the final protection.

Distributed locks alone should not be the only safeguard.

---

# 154. Worker Scaling

Because recurring posting is idempotent:

```text
1 worker
→ N workers
```

can later scale horizontally.

Database uniqueness prevents duplicate occurrence records.

---

# 155. API Scaling

API can eventually scale horizontally if:

```text
sessions are server-side in PostgreSQL

queue in Redis

uploads in MinIO

no local persistent state
```

---

# 156. Stateless API Principle

API process should avoid persistent local session state.

This supports:

```text
horizontal scaling

container restart

load balancing
```

---

# 157. Worker and API Shared Code

Both processes share:

```text
domain

services

repositories

finance engine

AI provider abstractions
```

Entry point differs.

Avoid separate copies of business logic.

---

# 158. Deployment Architecture — MVP

```text
Internet
   ↓
Reverse Proxy
   ↓
┌─────────────────────────────┐
│ Frontend                    │
│ Go API                      │
│ Go Worker                   │
│ PostgreSQL                  │
│ Redis                       │
│ MinIO                       │
└─────────────────────────────┘
```

For take-home:

```text
Docker Compose
```

is sufficient.

---

# 159. Production-Like Deployment

Potential:

```text
Reverse Proxy / TLS

Frontend Container

API Container(s)

Worker Container(s)

Managed / Container PostgreSQL

Redis

S3-compatible Object Storage
```

Kubernetes is not necessary for initial scope.

---

# 160. Graceful Shutdown

API and worker should support graceful shutdown.

On termination:

```text
stop accepting new work

finish / cancel active requests

close DB

close Redis

stop worker safely
```

Use Go context and signals.

---

# 161. Startup Flow

API startup:

```text
Load Config
    ↓
Validate Config
    ↓
Initialize Logger
    ↓
Connect PostgreSQL
    ↓
Connect Redis
    ↓
Initialize Repositories
    ↓
Initialize Services
    ↓
Initialize AI Provider
    ↓
Initialize Router
    ↓
Start HTTP Server
```

AI provider failure at startup should not necessarily prevent core startup if AI is optional.

---

# 162. Optional Dependency Initialization

If AI provider is misconfigured:

```text
AI marked unavailable
```

Core application may still start if:

```text
AI_ENABLED=false
```

or degraded mode is intentionally supported.

Missing required database configuration:

```text
startup fails
```

---

# 163. Worker Startup

```text
Load Config
    ↓
Connect PostgreSQL
    ↓
Connect Redis
    ↓
Initialize Services
    ↓
Initialize AI Provider
    ↓
Register Job Handlers
    ↓
Start Worker
```

---

# 164. Testing Architecture

Testing layers:

```text
Unit

Integration

API / Contract

Frontend Component

Frontend Integration

E2E
```

---

# 165. Backend Unit Tests

Highest value unit tests:

```text
financial formulas

forecast

scenario

budget

goal

recurring date calculation

AI output validation
```

---

# 166. Backend Integration Tests

Use real PostgreSQL for:

```text
transaction atomicity

row locking

migrations

foreign keys

unique constraints

repository queries

optimistic locking
```

SQLite should not replace PostgreSQL for behavior that depends on PostgreSQL semantics.

---

# 167. API Tests

Test:

```text
authentication

CSRF

ownership

validation

error mapping

financial response correctness
```

---

# 168. AI Tests

Use:

```text
MockProvider
```

Tests should not depend on live AI provider.

---

# 169. Frontend Tests

Use:

```text
Vitest

React Testing Library
```

Test:

```text
forms

interceptors

loading states

error states

scenario comparison

AI structured responses
```

---

# 170. Frontend API Mocking

Use:

```text
MSW
```

or equivalent for HTTP integration tests.

---

# 171. E2E

Potential:

```text
Playwright
```

Critical E2E flow:

```text
register

login

create account

add income

add expense

create budget

create goal

forecast

scenario

AI Copilot mock

logout
```

---

# 172. Test Clock

Finance calculations depend on time.

Inject a clock abstraction.

Example:

```go
type Clock interface {
    Now() time.Time
}
```

Production:

```text
RealClock
```

Test:

```text
FixedClock
```

---

# 173. Why Clock Abstraction

Without it:

```text
recurring dates

monthly periods

forecast horizons

goal calculations
```

become flaky around real current time.

---

# 174. Decimal Testing

All financial test cases should use exact decimal comparisons.

Avoid:

```text
0.1 + 0.2 floating-point assumptions
```

---

# 175. Migration Testing

CI should verify:

```text
empty database
↓
migrate up
↓
tests

migrate down
↓
migrate up
```

at least for latest migration behavior.

---

# 176. Security Testing

Test:

```text
cookie flags where possible

CSRF

refresh rotation

refresh replay

ownership

rate limit

session revocation

invalid UUID

resource enumeration
```

---

# 177. Concurrency Testing

Critical:

```text
concurrent expenses

concurrent transfers

refresh rotation race

optimistic locking

recurring duplicate worker processing
```

---

# 178. Error Testing

Every core module should test:

```text
not found

validation

conflict

unauthorized

unexpected repository error
```

---

# 179. CI Architecture

GitHub Actions may run:

```text
backend lint

backend test

frontend lint

frontend test

frontend build

migration test

Docker build
```

---

# 180. CI Environment

CI service containers may provide:

```text
PostgreSQL

Redis
```

MinIO only if file integration tests require it.

---

# 181. Linting

Backend:

```text
gofmt

go vet

golangci-lint
```

Frontend:

```text
ESLint

TypeScript check
```

---

# 182. Build Verification

Before merge/submission:

```text
go build

frontend build

docker compose config
```

must pass.

---

# 183. API Documentation

OpenAPI:

```text
docs/api/openapi.yaml
```

Swagger:

```text
/api/docs
```

The API contract should remain synchronized with implementation.

---

# 184. Architecture Documentation

Core architecture docs:

```text
product-foundation.md

business-requirements.md

user-flows.md

database-design.md

api-contract.md

ai-architecture.md

system-architecture.md
```

These are implementation references, not decorative documentation.

---

# 185. Architecture Decision Records — Optional

Important trade-offs may later use:

```text
docs/architecture/adr/
```

Examples:

```text
ADR-001 Modular Monolith

ADR-002 Opaque Refresh Tokens

ADR-003 Deterministic Finance Engine

ADR-004 AI Tool Architecture

ADR-005 Explicit SQL Migrations
```

Not mandatory if decisions are already clearly documented.

---

# 186. Scaling Strategy

Scale only after actual bottlenecks.

Order of likely optimization:

```text
indexes

query optimization

connection pool

background jobs

short-lived cache

multiple API instances

multiple workers

cursor pagination

advanced observability
```

---

# 187. Database Scaling

Potential future:

```text
vertical scaling

read replicas for analytics

partitioning for large transaction history
```

Not required for MVP.

---

# 188. Analytics Scaling

If analytics become expensive:

```text
materialized summaries

precomputed monthly aggregates

background snapshots
```

may be introduced.

Do not introduce prematurely.

---

# 189. AI Scaling

AI load can scale independently through:

```text
rate limits

worker concurrency

provider selection

background queue

context reduction
```

Copilot remains synchronous.

Insight generation can remain asynchronous.

---

# 190. AI Cost Scaling

Before using smaller/larger models dynamically:

```text
measure token usage
```

Potential future routing:

```text
categorization → cheap model

complex Copilot → stronger model
```

Not necessary initially.

---

# 191. Reliability Model

Savio separates dependencies by criticality.

## Critical

```text
PostgreSQL
```

Without it, Savio cannot serve financial functionality.

## Important but Degradable

```text
Redis
```

Depending on endpoint, core reads/writes may still work while async functionality degrades.

## Optional / Degradable

```text
AI Provider

MinIO when upload features unused
```

---

# 192. PostgreSQL Failure

Expected:

```text
readiness = false

financial API = unavailable
```

Return safe errors.

---

# 193. Redis Failure

Possible:

```text
API financial write still succeeds

background features degraded

rate limiting may fail closed or use local fallback depending on policy
```

For security-sensitive login rate limiting, fail behavior should be intentionally defined.

---

# 194. AI Failure

Expected:

```text
finance works

AI endpoints return degraded response

background AI jobs retry
```

---

# 195. MinIO Failure

If receipt upload is used:

```text
upload fails
```

but unrelated finance endpoints remain healthy.

---

# 196. Failure Isolation Principle

A failure in:

```text
AI
```

must not cascade into:

```text
transaction creation failure
```

unless the user explicitly requested an AI-only operation.

---

# 197. Recovery Strategy

Potential recovery mechanisms:

```text
queue retry

session cleanup

forecast recalculation

scenario recalculation

AI insight regeneration
```

Derived data should be reproducible from authoritative financial data.

---

# 198. Derived Data Principle

Data such as:

```text
forecast

scenario output

AI insight
```

may be recreated.

Data such as:

```text
transaction

transfer

account adjustment
```

must be treated as authoritative financial history.

---

# 199. Architecture Trade-Off — Persisted Account Balance

Decision:

```text
store current_balance
```

Benefit:

```text
fast account display
```

Cost:

```text
must maintain transactional correctness
```

Mitigation:

```text
balance changes only through financial services

row locking / atomic updates

reconciliation support

integration tests
```

---

# 200. Alternative Balance Model

Alternative:

```text
calculate balance from ledger every read
```

Benefit:

```text
single source from transaction history
```

Cost:

```text
potentially more expensive reads
```

Savio's persisted balance is acceptable if writes are carefully controlled.

---

# 201. Architecture Trade-Off — Forecast Snapshot

Decision:

```text
support persisted forecast snapshots
```

Benefits:

```text
reproducibility

scenario reference

freshness detection
```

Cost:

```text
derived-data lifecycle
```

---

# 202. Architecture Trade-Off — Offset Pagination

Decision:

```text
offset pagination for MVP
```

Benefits:

```text
simple UI

easy implementation
```

Cost:

```text
less efficient at very high offsets
```

Future:

```text
cursor pagination
```

---

# 203. Architecture Trade-Off — Redis Queue

Decision:

```text
Redis queue
```

Benefits:

```text
already useful for rate limiting

simple infrastructure

good fit for background jobs
```

Cost:

```text
less messaging flexibility than dedicated brokers
```

Sufficient for Savio's workload.

---

# 204. Architecture Trade-Off — No Microservices

Decision:

```text
modular monolith
```

Reason:

```text
strong transactional needs

small deployment

reviewability

lower complexity
```

Future extraction is possible if a module later requires independent scale.

---

# 205. Architecture Trade-Off — AI Provider Abstraction

Decision:

```text
provider interface
```

Benefit:

```text
testability

provider replacement

mock mode
```

Cost:

```text
small abstraction layer
```

Worth implementing because AI is an external dependency.

---

# 206. Architecture Trade-Off — No Generic Agent

Savio does not initially need a fully autonomous general-purpose AI agent.

Instead:

```text
bounded orchestration
+
explicit tools
```

This is safer and easier to explain.

---

# 207. Architecture Trade-Off — No Vector Database Initially

Savio does not require vector search for core financial operations.

Potential future semantic search may justify:

```text
pgvector
```

later.

Do not add it just because AI exists.

---

# 208. Architecture Trade-Off — No Event Bus Initially

Internal service calls remain synchronous.

Asynchronous side effects use Redis queue.

A full event bus is unnecessary at current scale.

---

# 209. Security Acceptance Criteria

System architecture is acceptable when:

```text
auth uses secure cookies

no auth tokens in browser storage

CSRF enforced

refresh rotates

session revoke works

financial resources are ownership-scoped

AI cannot bypass authorization

secrets remain server-side

sensitive logs are minimized

rate limiting exists
```

---

# 210. Financial Correctness Acceptance Criteria

```text
no floating-point finance math

balance updates atomic

transfer atomic

correction uses void + replacement

transaction voiding restores effect

recurring posting idempotent

concurrent writes do not lose updates

forecast is deterministic

scenario does not modify real data
```

---

# 211. AI Acceptance Criteria

```text
AI is optional

AI provider mock exists

AI uses deterministic finance tools

AI output is validated

AI cannot mutate finance state silently

AI context is minimized

AI failure is isolated

AI actions are allowlisted
```

---

# 212. Infrastructure Acceptance Criteria

```text
docker compose starts dependencies

PostgreSQL migration works from empty DB

Redis reachable

MinIO reachable if enabled

API health endpoint works

worker starts

frontend builds

backend builds
```

---

# 213. Observability Acceptance Criteria

```text
request IDs

structured logs

safe error codes

worker logs

AI request metadata

health

readiness
```

---

# 214. System Request Flow — Financial Write

Example expense creation:

```text
React Form
    ↓
Axios
    ↓
CSRF Header + Cookies
    ↓
Go API
    ↓
Request ID
    ↓
Auth
    ↓
CSRF
    ↓
Validation
    ↓
Transaction Handler
    ↓
Transaction Service
    ↓
Authorization / Ownership
    ↓
DB Transaction
        ├── lock account
        ├── create transaction
        ├── update balance
        ├── stale forecast
        ├── stale scenarios
        └── audit
    ↓
COMMIT
    ↓
Queue optional signal job
    ↓
API Response
    ↓
TanStack Query Invalidation
    ↓
Updated UI
```

---

# 215. System Request Flow — Forecast

```text
React Forecast Page
    ↓
POST /forecast/calculate
    ↓
Forecast Handler
    ↓
Forecast Service
    ↓
Load:
Accounts
Recurring
History
Assumptions
    ↓
Forecast Engine
    ↓
Deterministic Result
    ↓
Optional Snapshot Persist
    ↓
Response
    ↓
UI Timeline
```

No AI is required.

---

# 216. System Request Flow — Scenario

```text
Scenario UI
    ↓
Scenario Definition
    ↓
POST /scenarios/:id/calculate
    ↓
Scenario Service
    ↓
Build Baseline
    ↓
Apply Modifications
    ↓
Scenario Engine
    ↓
Comparison
    ↓
Persist Snapshot
    ↓
Response
```

Then separately:

```text
Explain Scenario
    ↓
AI Orchestrator
    ↓
Load Stored Scenario Snapshot
    ↓
AI Provider
    ↓
Validated Explanation
```

---

# 217. System Request Flow — Copilot

```text
User Question
    ↓
React Copilot
    ↓
POST /ai/copilot
    ↓
Auth + Rate Limit
    ↓
AI Orchestrator
    ↓
Intent
    ↓
Finance Tools
    ↓
Finance Services
    ↓
PostgreSQL
    ↓
Structured Facts
    ↓
Context Builder
    ↓
AI Provider
    ↓
Structured Output Validation
    ↓
Response
```

---

# 218. System Background Flow — Recurring Posting

```text
Scheduled Worker
    ↓
Find Due Rules
    ↓
For Each Rule
    ↓
Recurring Service
    ↓
BEGIN
    ↓
Create unique occurrence
    ↓
Create transaction
    ↓
Update account balance
    ↓
Update next occurrence
    ↓
Audit
    ↓
COMMIT
```

Duplicate worker attempt:

```text
unique occurrence conflict
→ safely skip
```

---

# 219. System Background Flow — AI Insight

```text
Financial Change
    ↓
Signal Evaluation
    ↓
Meaningful Signal?
    ├── No → End
    └── Yes
        ↓
Queue Job
        ↓
Worker
        ↓
Dedup Check
        ↓
Build Context
        ↓
AI Provider
        ↓
Validate
        ↓
Store Insight
        ↓
Notification
```

---

# 220. Production Data Flow Hierarchy

```text
USER INPUT
    ↓
HTTP VALIDATION
    ↓
AUTHORIZATION
    ↓
BUSINESS RULES
    ↓
AUTHORITATIVE FINANCIAL RECORD
    ↓
DETERMINISTIC CALCULATION
    ↓
DERIVED FINANCIAL STATE
    ↓
OPTIONAL AI INTERPRETATION
    ↓
UI
```

Never:

```text
AI OUTPUT
    ↓
DIRECT FINANCIAL RECORD
```

without validated user-confirmed action.

---

# 221. System Dependency Hierarchy

```text
Platform Infrastructure
        ↓
Repository
        ↓
Domain / Finance Engine
        ↓
Application Services
        ↓
HTTP / Worker Interfaces
```

AI:

```text
Finance Services
        ↓
AI Tools
        ↓
AI Orchestrator
```

Frontend:

```text
REST Contract
        ↓
React Features
```

---

# 222. Module Dependency Rules

Recommended:

```text
transactions
→ accounts
→ categories

budgets
→ transactions
→ categories

analytics
→ transactions

forecast
→ accounts
→ recurring
→ analytics

scenarios
→ forecast
→ goals

ai
→ analytics
→ forecast
→ scenarios
```

Avoid circular dependencies.

---

# 223. Cross-Module Communication

Prefer calling:

```text
application service interfaces
```

rather than reaching into another module's repository directly.

Example:

AI:

```text
AnalyticsService.GetPeriodComparison()
```

rather than:

```text
TransactionRepository custom AI query
```

when the domain operation already exists.

---

# 224. Service Interface Example

```go
type CashflowReader interface {
    GetSummary(
        ctx context.Context,
        userID uuid.UUID,
        period DateRange,
    ) (CashflowSummary, error)
}
```

AI tools can depend on narrow interfaces.

---

# 225. Avoid God Service

Do not create:

```text
FinanceService
```

with hundreds of unrelated methods.

Separate:

```text
TransactionService

AnalyticsService

ForecastService

ScenarioService
```

and share domain calculation packages.

---

# 226. Avoid Over-Abstracting

Do not create generic:

```text
BaseRepository[T]

BaseService[T]

GenericController[T]
```

if they obscure domain rules.

Financial workflows benefit from explicit domain code.

---

# 227. Architecture Principle — Explicit Over Clever

Prefer:

```text
TransferService.Create()
```

with explicit transaction logic over overly generic CRUD abstractions.

---

# 228. Architecture Principle — Domain Names

Use Savio terminology consistently:

```text
Account

Transaction

Transfer

Recurring Transaction

Budget

Financial Goal

Forecast

Scenario

AI Insight
```

Avoid ambiguous internal names if business terms already exist.

---

# 229. Architecture Principle — Determinism

Any value displayed as authoritative should be reproducible from:

```text
database state
+
explicit configuration
+
calculation version
```

AI output is an exception because it is interpretive, not authoritative.

---

# 230. Architecture Principle — Reproducibility

Forecast and scenario snapshots may record:

```text
calculation version

data through date

assumptions

generated time
```

so results can be explained later.

---

# 231. Architecture Principle — Safe Degradation

Savio should degrade by capability.

Example:

```text
AI down
→ AI unavailable
→ finance works
```

Not:

```text
AI down
→ entire application unavailable
```

---

# 232. Architecture Principle — Privacy by Design

Private financial data should be scoped and minimized at each boundary.

```text
DB query
→ only user data

API
→ only required DTO

AI
→ only relevant context

logs
→ metadata, not full sensitive payload
```

---

# 233. Architecture Principle — Testability

Critical domain behavior should be testable without:

```text
browser

live AI provider

production Redis

external network
```

Finance engine tests should be pure/deterministic where possible.

---

# 234. Architecture Principle — Reviewer Simplicity

A reviewer should be able to understand:

```text
where auth lives

where financial rules live

where AI lives

where database queries live

where background jobs live
```

without reading the whole repository.

---

# 235. Initial Deployment Recommendation

For the take-home:

```text
Docker Compose
```

with:

```text
frontend
api
worker
postgres
redis
minio
```

This provides a reproducible environment while staying understandable.

---

# 236. Explicitly Avoided Architecture

Do not add without demonstrated need:

```text
Kubernetes

Kafka

RabbitMQ

Elasticsearch

GraphQL

multiple backend microservices

service mesh

full distributed event sourcing

CQRS framework

multiple AI agents

vector database

complex model router
```

---

# 237. Event Sourcing Decision

Savio does not use full event sourcing.

Financial history is explicit through:

```text
transactions

transfers

adjustments

audit logs
```

A full event-sourced architecture would add unnecessary complexity.

---

# 238. CQRS Decision

Savio may naturally have:

```text
write services

read analytics
```

but does not require a formal CQRS framework.

Use simple separated queries where helpful.

---

# 239. Domain Events

Internal domain events may be introduced later.

Example:

```text
TransactionCreated

BudgetThresholdReached

ForecastBecameStale
```

For MVP, direct service orchestration plus async jobs is enough.

---

# 240. Future Outbox Event Example

```json
{
  "type": "TRANSACTION_CREATED",
  "aggregate_id": "transaction-uuid",
  "user_id": "user-uuid",
  "occurred_at": "2026-08-24T15:00:00Z"
}
```

This can trigger:

```text
signal evaluation

notification

analytics refresh
```

later.

---

# 241. Future Household Architecture

If household finance is introduced:

```text
Household
    ↓
Membership
    ↓
Role / Permission
    ↓
Shared Resource Scope
```

Authorization would evolve from:

```text
resource.user_id
```

to:

```text
resource ownership scope
```

This is intentionally deferred.

---

# 242. Future Multi-Currency Architecture

Future multi-currency would require:

```text
currency-aware accounts

FX rates

base currency

conversion date

rate source

realized/unrealized conversion rules
```

This is intentionally outside MVP to preserve financial correctness.

---

# 243. Future Bank Integration

Bank integration would add:

```text
external provider credentials

webhooks

sync jobs

transaction matching

deduplication

consent lifecycle
```

This is not required for the initial Savio architecture.

---

# 244. Future Receipt Architecture

```text
Receipt Upload
    ↓
MinIO
    ↓
Async Extraction
    ↓
Structured Draft
    ↓
User Review
    ↓
Normal Transaction Service
```

The normal financial service remains the final write authority.

---

# 245. Future Import Architecture

```text
File Upload
    ↓
Object Storage
    ↓
Import Job
    ↓
Parser
    ↓
Validation
    ↓
Review
    ↓
Confirmed Rows
    ↓
Transaction Service
```

Never import directly into database bypassing business services.

---

# 246. Architecture Review Checklist

Before implementation, each new module should answer:

```text
What domain problem does it solve?

What data does it own?

What service owns business rules?

What repository owns persistence?

Does it affect financial balance?

Does it require a transaction?

Does it require optimistic locking?

Does it invalidate forecast?

Does it invalidate scenario?

Does it create audit events?

Does it enqueue work?

Does AI participate?

Can AI failure be isolated?

How is ownership enforced?

What tests prove correctness?
```

---

# 247. Architecture Definition of Done

A system-level feature is complete when:

```text
domain rule implemented

API contract satisfied

database integrity preserved

authorization enforced

validation implemented

error handling implemented

tests added

logging exists

UI state handled

documentation updated
```

For async features:

```text
idempotency tested

retry behavior defined

failure behavior defined
```

For AI features:

```text
structured context

validated output

provider failure handling

human control
```

---

# 248. Critical Architecture Demo

A technical demo should be able to explain this flow clearly:

```text
Create Expense
    ↓
Cookie Authentication
    ↓
CSRF Validation
    ↓
Backend Validation
    ↓
Transaction Service
    ↓
Database Transaction
    ↓
Account Balance Updated Safely
    ↓
Forecast Marked Stale
    ↓
Audit Created
    ↓
Async Signal Processing
    ↓
AI Insight Generated
    ↓
Frontend Refreshes
```

This demonstrates multiple architecture decisions through one business action.

---

# 249. Critical Failure Demo

Useful technical scenarios:

```text
Concurrent balance updates
→ no lost update

Duplicate recurring job
→ one transaction

AI provider down
→ finance remains functional

Expired access token
→ one refresh request

Stale budget update
→ 409

Cross-user transaction access
→ denied

Database transaction failure during transfer
→ no partial balance movement
```

---

# 250. Architecture Summary

Savio's architecture can be summarized as:

```text
React Frontend
      ↓
REST API
      ↓
Go Modular Monolith
      │
      ├── Authentication & Security
      │
      ├── Financial Domain Services
      │
      ├── Deterministic Finance Engine
      │
      ├── Forecast & Scenario Engine
      │
      ├── AI Orchestration
      │
      └── Background Job Integration
      │
      ├───────────────┬───────────────┬───────────────┐
      ▼               ▼               ▼               ▼
 PostgreSQL         Redis           MinIO         AI Provider
      ▲               │
      │               ▼
      │           Go Worker
      └───────────────┘
```

---

# 251. Source-of-Truth Hierarchy

The architecture obeys the following hierarchy:

```text
SECURITY & AUTHORIZATION
        ↓
AUTHORITATIVE FINANCIAL DATA
        ↓
BUSINESS RULES
        ↓
DETERMINISTIC FINANCE ENGINE
        ↓
DERIVED FINANCIAL RESULTS
        ↓
AI INTERPRETATION
        ↓
USER DECISION
```

A lower layer must never override a higher layer.

---

# 252. Final Architecture Principle

The Savio architecture should remain:

```text
secure

deterministic

modular

testable

auditable

explainable

failure-tolerant
```

without becoming unnecessarily distributed or complex.

The system must preserve three boundaries:

> **PostgreSQL stores authoritative financial state.**

> **The Finance Engine produces authoritative financial calculations.**

> **AI provides interpretation and assistance, not financial authority.**

The final product rule remains:

> **Finance Engine calculates. AI interprets. User decides.**
# Savio — AGENTS.md

## 1. Purpose

This file defines the working rules for AI coding agents and automated engineering assistants operating inside the Savio repository.

Its purpose is to ensure that generated code remains aligned with Savio's:

```text
product requirements
business rules
architecture
security model
database design
API contract
frontend architecture
testing strategy
implementation plan
```

This file is not a replacement for the project documentation.

It is an execution guide that tells coding agents:

```text
what to read
what to trust
what they may change
what they must not assume
how implementation should be verified
```

The central product rule is:

> **Finance Engine calculates. AI interprets. User decides.**

Every implementation decision must preserve this rule.

---

# 2. Project Identity

Project:

```text
Savio
```

Category:

```text
Personal Cashflow Intelligence & Financial Decision Support Platform
```

Repository:

```text
savio
```

Core user journey:

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

Primary product pillars:

```text
UNDERSTAND

PREDICT

SIMULATE

EXPLAIN
```

---

# 3. Core Product Thesis

Savio helps users:

```text
understand their current financial condition

identify meaningful cashflow patterns

forecast what may happen next

simulate financial decisions before making them

understand trade-offs through AI-assisted explanation
```

Savio must not become:

```text
a generic CRUD finance tracker

an expense tracker with a chatbot added

a trading application

an investment advisory platform

an autonomous financial decision-maker
```

---

# 4. Mandatory Source-of-Truth Documents

Before implementing a feature, read the relevant documentation.

Primary sources:

```text
README.md

DESIGN.md

docs/product/product-foundation.md

docs/product/business-requirements.md

docs/product/user-flows.md

docs/database/database-design.md

docs/api/api-contract.md

docs/architecture/system-architecture.md

docs/architecture/frontend-architecture.md

docs/architecture/ai-architecture.md

docs/engineering/security.md

docs/engineering/testing-strategy.md

docs/engineering/implementation-plan.md

docs/assignment/take-home-test-specification.md
```

---

# 5. Documentation Precedence

If documents appear to conflict, use this precedence order:

```text
1. Explicit current task requirements
2. docs/product/business-requirements.md
3. docs/api/api-contract.md
4. docs/database/database-design.md
5. docs/engineering/security.md
6. docs/architecture/system-architecture.md
7. docs/architecture/ai-architecture.md
8. docs/architecture/frontend-architecture.md
9. DESIGN.md
10. docs/engineering/testing-strategy.md
11. docs/engineering/implementation-plan.md
12. README.md
```

However:

> **Do not silently resolve contradictions.**

If two source-of-truth documents materially disagree:

```text
stop
↓
identify the conflict
↓
propose the smallest coherent resolution
↓
update the affected documentation with the implementation
```

---

# 6. Assignment Constraints

The implementation must continue satisfying the take-home requirements.

Important constraints include:

```text
Go backend

Gin

GORM

PostgreSQL

REST API

/api/v1 versioning

React or Vue frontend

TypeScript preferred

Tailwind CSS

Axios

cookie authentication

no authentication tokens in localStorage

no authentication tokens in sessionStorage

CSRF protection

meaningful authorization levels

backend-enforced authorization

database migrations

rollback support

validation

search

filter

sort

pagination

good UI / UX

maintainable architecture
```

Bonus engineering is valuable only after core correctness.

---

# 7. Approved Technology Direction

Backend:

```text
Go
Gin
GORM
PostgreSQL
golang-migrate
decimal-safe money library
Redis
MinIO
```

Frontend:

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

Testing:

```text
Go testing
Testify
Vitest
React Testing Library
MSW
Playwright
```

AI:

```text
provider abstraction
OpenAI-compatible provider
mock provider
structured output
bounded finance tools
```

---

# 8. Architecture Style

Savio is a:

```text
MODULAR MONOLITH
```

with:

```text
Go API process

Go background worker

PostgreSQL

Redis

MinIO

external AI provider
```

Do not convert Savio into microservices without explicit architectural approval.

---

# 9. Backend Dependency Direction

The preferred backend dependency flow is:

```text
HTTP Handler
     ↓
Application Service
     ↓
Domain / Finance Engine
     ↓
Repository
     ↓
PostgreSQL
```

Supporting infrastructure:

```text
platform
config
logger
security
queue
storage
```

---

# 10. AI Dependency Direction

Correct:

```text
AI Orchestrator
      ↓
AI Tools
      ↓
Finance Services
      ↓
Finance Engine
```

Forbidden:

```text
Finance Engine
      ↓
AI Provider
```

Core financial calculations must never depend on AI availability.

---

# 11. Frontend Dependency Direction

Preferred:

```text
Page
 ↓
Feature Component / Hook
 ↓
Feature API Module
 ↓
Shared Axios Client
 ↓
Backend
```

Avoid:

```text
random Axios calls directly inside UI components
```

---

# 12. Backend Module Boundaries

Expected modules include:

```text
auth

sessions

users

workspaces

settings

accounts

categories

transactions

transfers

recurring

budgets

goals

analytics

forecast

scenarios

ai

notifications

audit
```

Shared infrastructure belongs in:

```text
internal/platform
```

---

# 13. Handler Rules

Handlers may:

```text
bind request input

extract path/query parameters

read authenticated context

invoke validation

call service

map service result to API response
```

Handlers must not contain:

```text
complex business rules

financial formulas

large database workflows

authorization logic duplicated from services

AI orchestration internals
```

---

# 14. Service Rules

Services own:

```text
business rules

authorization

resource scoping

state transitions

transaction boundaries

cross-module orchestration
```

Example:

```text
TransactionService.CreateExpense()
```

is preferable to generic CRUD service behavior.

---

# 15. Repository Rules

Repositories own:

```text
database reads

database writes

aggregations

row locking

query construction
```

Repositories should not become the primary place for business policy.

---

# 16. Generic Abstraction Rule

Do not create abstractions only to reduce line count.

Avoid unnecessary:

```text
BaseController

BaseService

GenericCRUDService

GenericRepository[T]
```

when they obscure domain rules.

Prefer explicit domain behavior.

---

# 17. Financial Authority

Authoritative financial data comes from:

```text
PostgreSQL financial records
```

Financial calculations come from:

```text
deterministic finance services / engines
```

AI output is:

```text
interpretive
```

not authoritative.

---

# 18. Money Rules

Never use floating-point types for authoritative money calculations.

Forbidden:

```go
float32
float64
```

for financial amounts.

Use:

```text
PostgreSQL BIGINT

integer minor units

ISO currency code
```

Money arithmetic uses a decimal-safe representation over integer minor units.

Forbidden:

```text
floating point
```

for authoritative amounts.

API monetary values use decimal-safe strings, converted to/from minor units.

Example:

```json
{
  "amount": "1500000.00"
}
```

represents 150,000,000 minor units at 2-decimal scale.

---

# 19. Currency Rule

P0 uses one base currency per workspace.

Do not silently implement FX conversion.

If an account currency differs from workspace base currency and multi-currency has not been explicitly implemented:

```text
reject the operation
```

or follow the documented business rule.

---

# 20. Time Rules

Persist timestamps in:

```text
UTC
```

Business date behavior should respect:

```text
workspace timezone
```

Date-only financial fields should not accidentally shift due to timezone conversion.

---

# 21. Workspace Model

Every user has a default personal workspace.

Financial resources are scoped by:

```text
workspace_id
```

Membership roles:

```text
OWNER

MEMBER

VIEWER
```

---

# 22. Workspace Authorization

All workspace resources must verify:

```text
authenticated user
+
active workspace membership
+
required permission
```

Never trust:

```text
workspace_id from client
```

without membership verification.

---

# 23. Role Semantics

OWNER:

```text
read finance

write finance

manage members

change roles

manage workspace
```

MEMBER:

```text
read finance

write finance
```

VIEWER:

```text
read-only
```

---

# 24. Last Owner Invariant

Never allow:

```text
removal of the final OWNER
```

or:

```text
demotion of the final OWNER
```

without a valid ownership transfer.

---

# 25. Resource Ownership / Scope

For every resource lookup:

```text
resource ID alone is never authorization
```

Queries must be workspace-scoped.

Bad:

```text
FindTransactionByID(id)
```

Preferred:

```text
FindTransactionByID(workspaceID, id)
```

---

# 26. IDOR Rule

Never implement:

```text
SELECT resource WHERE id = ?
```

for protected user-facing resources without authorization scoping.

Cross-workspace identifiers must not expose data.

---

# 27. Account Rules

Accounts represent:

```text
cash

bank accounts

e-wallets

savings accounts
```

P0 does not model:

```text
credit cards as borrowing engine

investment portfolios

crypto wallets with market valuation
```

unless explicitly added later.

---

# 28. Account Balance Rule

Preferred financial source of truth:

```text
opening balance
+
posted ledger effects
```

Do not introduce an independently mutable current balance.

If a cached balance is added for performance:

```text
it is derived state
```

and must be reconstructable.

---

# 29. Account Reconciliation

Never directly overwrite financial history to fix a balance.

Use:

```text
ADJUSTMENT
```

record.

Example:

```text
tracked balance:
4.8M

actual:
5.0M

adjustment:
+200k
```

---

# 30. Account Archiving

If an account has financial history:

```text
archive
```

rather than hard delete.

Archived accounts remain available for historical display.

Archived accounts must not accept new ordinary financial activity.

---

# 31. Category Rules

Categories have a type:

```text
INCOME

EXPENSE
```

An income transaction cannot use an expense category.

An expense transaction cannot use an income category.

System categories may be globally available.

Custom categories must remain workspace-scoped.

---

# 32. Transaction Types

Core transaction types:

```text
INCOME

EXPENSE

ADJUSTMENT
```

Transfers are modeled separately.

---

# 33. Transaction Status

Canonical P0 lifecycle:

```text
DRAFT
→ POSTED
→ VOIDED
```

A transaction is created as `DRAFT`, becomes `POSTED` when it takes financial effect, and is `VOIDED` to invalidate it.

`POSTED` financial fields (amount, account, type, date) are immutable.

Correction is performed by:

```text
VOID the original
→ create a replacement transaction
```

not by destructive historical editing.

---

# 34. Transaction Amount Rule

Amounts should normally be stored as:

```text
positive values
```

Direction comes from transaction type.

Avoid ambiguous combinations such as:

```text
EXPENSE
amount = -100000
```

unless documentation explicitly defines that model.

---

# 35. Transaction Analytics Rule

Normal income/expense analytics should include:

```text
POSTED INCOME

POSTED EXPENSE
```

Exclude:

```text
VOIDED

TRANSFER

ordinary ADJUSTMENT
```

unless an endpoint explicitly asks for adjustment activity.

---

# 36. Transaction Editing

A still-pending `DRAFT` transaction may be edited before it is posted.

`POSTED` transactions are financially immutable; their financial fields must not be silently rewritten.

To change a posted transaction:

```text
void the original
→ create a replacement transaction
→ preserve audit history
→ recompute authoritative derived results
```

Never overwrite without optimistic concurrency protection.

---

# 37. Transaction Voiding

Voiding must:

```text
preserve historical record

mark transaction VOIDED

exclude it from active analytics

recalculate derived balance
```

A VOIDED transaction must not be voided twice.

---

# 38. Transfer Rules

Transfer:

```text
source account
→ destination account
```

must satisfy:

```text
different accounts

same workspace

supported currency compatibility

active accounts

amount > 0
```

---

# 39. Transfer Invariant

A transfer must not change total portfolio balance.

Example:

```text
before:
A + B = 10M

after:
A + B = 10M
```

Transfers must not count as income or expense.

---

# 40. Transfer Atomicity

Transfer creation and voiding must be atomic.

No state may exist where:

```text
source changed

destination did not
```

after successful response.

---

# 41. Recurring Transaction Philosophy

Recurring transactions represent:

```text
planned / expected financial activity
```

They are not automatically equivalent to actual ledger history.

---

# 42. Recurring Default

Preferred:

```text
auto_post = false
```

by default.

Occurrences begin as:

```text
PENDING
```

and become actual when:

```text
CONFIRMED
```

or auto-posted when explicitly configured.

---

# 43. Recurring Status

Preferred:

```text
ACTIVE

PAUSED

ENDED
```

Do not introduce new status names without updating all contracts.

---

# 44. Recurring Occurrence Status

Preferred:

```text
PENDING

CONFIRMED

SKIPPED
```

Optional:

```text
FAILED
```

for auto-post infrastructure.

---

# 45. Recurring Idempotency

A recurring occurrence may become actual at most once.

Final protection should be a database unique constraint.

Do not rely only on:

```text
Redis locks

worker memory

frontend disabled buttons
```

---

# 46. Budget Scope

P0 budgets are intentionally simple.

Default model:

```text
monthly

expense category

amount
```

One active category budget for a given period.

---

# 47. Budget Persistence Status

Recommended:

```text
ACTIVE

CLOSED
```

---

# 48. Budget Computed Status

Computed from deterministic values:

```text
ON_TRACK

WARNING

EXCEEDED
```

Do not store these as authoritative if they can be derived reliably.

---

# 49. Budget Actual Spend

Budget actual is derived from:

```text
POSTED EXPENSE transactions

matching category

matching period
```

Exclude:

```text
transfers

voided transactions

adjustments
```

---

# 50. Goals Scope

P0 financial goals are lightweight planning entities.

Core fields:

```text
target amount

tracked current amount

target date

priority
```

Do not claim that goal progress automatically represents money physically reserved in an account unless such allocation mechanics are actually implemented.

---

# 51. Goal Calculations

Deterministic calculations include:

```text
progress percentage

remaining amount

required monthly contribution

simple feasibility
```

AI may explain them.

AI must not calculate them as authoritative values.

---

# 52. Analytics Rules

Analytics must be deterministic.

Use PostgreSQL aggregation where appropriate.

Do not fetch entire transaction history into Go or frontend for simple:

```text
SUM

GROUP BY

period comparison
```

operations.

---

# 53. Frontend Analytics Rule

Frontend must not independently calculate authoritative:

```text
income

expense

net cashflow

savings rate

budget utilization
```

from raw transactions when backend endpoints already provide them.

---

# 54. Forecast Philosophy

Forecast is:

```text
deterministic

assumption-driven

explainable
```

It is not an LLM prediction.

Default horizon:

```text
90 days
```

---

# 55. Forecast Inputs

Typical inputs:

```text
current derived liquid balance

recurring income

recurring expenses

planned occurrences

historical variable spending

explicit assumptions
```

---

# 56. Forecast Outputs

Typical outputs:

```text
opening balance

ending balance

minimum balance

minimum-balance date

projected income

projected expense

timeline

confidence

assumptions

calculation version
```

---

# 57. Forecast Event Types

Use one normalized vocabulary.

Preferred:

```text
KNOWN

SCHEDULED

ESTIMATED

ASSUMED
```

KNOWN: explicitly known future events.

SCHEDULED: generated from recurring rules.

ESTIMATED: derived from historical behavior.

ASSUMED: explicit user assumptions.

---

# 58. Forecast Confidence

Forecast confidence is:

```text
deterministic metadata
```

based on data quality and assumptions.

It is not AI confidence.

Do not generate confidence by asking the LLM.

---

# 59. Forecast Freshness

Relevant financial changes should invalidate existing forecast snapshots.

Use:

```text
FRESH

STALE
```

or documented equivalent.

Frontend cache freshness and domain forecast freshness are separate concepts.

---

# 60. Scenario Philosophy

Scenario is:

> **A non-destructive overlay over a deterministic baseline forecast.**

Scenario calculations must never mutate actual financial records.

---

# 61. Scenario Types

P0 may support:

```text
ONE_TIME_EXPENSE

ONE_TIME_INCOME

RECURRING_EXPENSE

RECURRING_INCOME

INCOME_REDUCTION

INCOME_REMOVAL

EXPENSE_REDUCTION
```

Do not add arbitrary scenario types without updating business rules and tests.

---

# 62. Scenario Invariant

After calculating a scenario:

```text
real accounts unchanged

real transactions unchanged

real recurring rules unchanged

real budgets unchanged
```

Only:

```text
scenario state

scenario snapshot
```

may be persisted.

---

# 63. Scenario Freshness

If source finance data changes after calculation:

```text
scenario becomes stale
```

Do not silently present stale scenario output as current.

---

# 64. AI Core Rule

AI is an:

```text
untrusted probabilistic subsystem
```

It may:

```text
explain

summarize

classify

suggest

help interpret
```

It may not become authoritative for financial state.

---

# 65. AI Allowed Responsibilities

Examples:

```text
transaction category suggestion

financial signal explanation

forecast explanation

scenario explanation

grounded natural-language finance questions
```

---

# 66. AI Forbidden Responsibilities

Never allow AI to directly determine authoritative:

```text
account balance

income totals

expense totals

budget utilization

savings rate

forecast output

scenario calculations

goal progress
```

---

# 67. AI Write Rule

AI must not silently perform financial mutations.

Future write flow must be:

```text
AI proposal
↓
backend validation
↓
user confirmation
↓
normal domain service
```

---

# 68. AI Tools

AI accesses financial facts only through bounded domain tools.

Allowed examples:

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

# 69. Forbidden AI Tools

Do not expose:

```text
execute_sql

shell

filesystem

arbitrary HTTP request

database admin

arbitrary user lookup
```

---

# 70. AI Authorization Context

Authenticated context is injected by backend.

The LLM must never choose:

```text
user_id

workspace_id

authorization role
```

as a trusted value.

---

# 71. AI Resource IDs

Any AI-provided resource ID is untrusted.

Example:

```text
goal_id
```

must still pass normal workspace authorization.

---

# 72. AI Context Minimization

Send:

```text
only relevant financial data
```

Prefer:

```text
aggregates
```

before:

```text
raw transaction history
```

---

# 73. AI Secret Rule

Never send to the model:

```text
password

password hash

access token

refresh token

CSRF token

session secret

database credential

API key
```

---

# 74. AI Prompt Injection

Treat:

```text
transaction descriptions

merchant names

uploaded extracted text

user-entered notes
```

as untrusted data.

Never allow instructions embedded inside those fields to expand AI capabilities.

---

# 75. AI Output Validation

Structured AI output must be schema validated before:

```text
persistence

frontend rendering as application structure

action execution
```

Unknown actions or enums must be rejected.

---

# 76. AI Provider Abstraction

Business logic should depend on an interface.

Production:

```text
real provider adapter
```

Tests:

```text
mock provider
```

Do not call live AI in automated CI tests.

---

# 77. AI Failure Isolation

If AI is unavailable:

```text
accounts still work

transactions still work

analytics still work

budgets still work

forecast still works

scenario calculation still works
```

AI endpoints may degrade gracefully.

---

# 78. Authentication Rule

Authentication is cookie-based.

Do not introduce:

```text
localStorage access token

sessionStorage access token
```

---

# 79. Access Credential

Recommended:

```text
short-lived JWT
```

stored as:

```text
HttpOnly cookie
```

---

# 80. Refresh Credential

Recommended:

```text
opaque random refresh token
```

stored as:

```text
HttpOnly cookie
```

Database stores:

```text
hash
```

not raw token.

---

# 81. Refresh Rotation

Every successful refresh rotates the token.

Old token becomes invalid.

Backend concurrency protection is mandatory.

Frontend single-flight refresh does not replace backend rotation safety.

---

# 82. CSRF Rule

All state-changing cookie-authenticated requests must follow the documented CSRF strategy.

Protected methods:

```text
POST

PUT

PATCH

DELETE
```

Do not disable CSRF to simplify development.

---

# 83. Cookie Security

Production cookies must use appropriate:

```text
HttpOnly

Secure

SameSite

Path

Max-Age / Expires
```

Do not weaken production settings for local convenience.

Use environment-specific configuration.

---

# 84. Password Rule

Use:

```text
Argon2id
```

or approved secure password hashing.

Never implement custom password cryptography.

---

# 85. Rate Limiting

At minimum protect:

```text
login

register

refresh

AI Copilot

AI categorization
```

Use server-side rate limiting.

---

# 86. Validation

Backend validates:

```text
body

path

query

enum

UUID

dates

money

pagination

sort

resource relationships
```

Frontend validation is UX only.

---

# 87. Sort Security

Never concatenate raw sort fields into SQL.

Use:

```text
allowlist mapping
```

---

# 88. SQL Rule

Always use parameterized queries / safe GORM binding.

Never accept raw SQL from:

```text
client

AI

query parameter
```

---

# 89. Error Response Rule

Use centralized application errors.

Never return raw:

```text
err.Error()

SQL error

stack trace

panic

provider response containing secrets
```

to the user.

---

# 90. Standard API Success

Follow API contract.

Typical:

```json
{
  "success": true,
  "data": {}
}
```

---

# 91. Standard API Error

Typical:

```json
{
  "success": false,
  "error": {
    "code": "..."
  },
  "message": "..."
}
```

Field errors may include:

```text
error.details
```

---

# 92. HTTP Status Semantics

Use documented statuses consistently:

```text
200
201
204

400
401
403
404
409
422
429
500
503
```

Do not return:

```text
200
```

for business failures just to simplify frontend handling.

---

# 93. API Versioning

Public application endpoints use:

```text
/api/v1
```

Do not add unversioned feature endpoints casually.

Infrastructure endpoints may remain:

```text
/health

/ready
```

---

# 94. API Contract Update Rule

When changing:

```text
request schema

response schema

endpoint

error code
```

update:

```text
docs/api/api-contract.md

frontend types/tests
```

in the same task.

---

# 95. Database Migration Rule

Schema changes require explicit migration files.

Do not rely on:

```text
GORM AutoMigrate
```

as production schema source of truth.

Every migration should include:

```text
up

down
```

where reasonably reversible.

---

# 96. Database Constraint Rule

Use database constraints for simple invariants.

Examples:

```text
foreign keys

unique membership

unique recurring occurrence

positive amount

valid structural relationships
```

Do not rely solely on application checks when a database constraint can protect integrity.

---

# 97. Index Rule

Add indexes intentionally for:

```text
foreign keys

workspace-scoped listing

date range queries

recurring due lookup

frequent analytics filters
```

Do not add indexes blindly.

---

# 98. Database Performance Rule

For expensive queries:

```text
measure first
```

Use:

```text
EXPLAIN ANALYZE
```

for high-value query review.

---

# 99. Transaction Boundary Rule

Use explicit database transactions for operations that must be atomic.

Examples:

```text
registration workspace bootstrap

transfer

reconciliation

recurring occurrence confirmation

refresh token rotation
```

---

# 100. Concurrency Rule

Always consider concurrency for:

```text
financial writes

refresh rotation

recurring jobs

versioned edits
```

Do not assume requests execute one at a time.

---

# 101. Optimistic Locking

Use:

```text
version
```

on mutable resources where stale form submissions are possible.

Typical conflict:

```text
409 VERSION_CONFLICT
```

---

# 102. Worker Rule

Background worker must call shared services.

Do not duplicate domain behavior in job handlers.

Preferred:

```text
Worker Job
↓
Service
```

not:

```text
Worker Job
↓
custom duplicate finance logic
```

---

# 103. Queue Rule

Queue payloads should contain minimal references.

Preferred:

```json
{
  "workspace_id": "...",
  "resource_id": "..."
}
```

rather than large financial snapshots.

Reload authoritative state during processing.

---

# 104. Queue Failure

Non-critical async failure must not corrupt already committed financial state.

Example:

```text
transaction committed

AI insight enqueue failed
```

The transaction remains valid.

---

# 105. Worker Idempotency

Every retryable financial or notification job must define an idempotency strategy.

Especially:

```text
recurring occurrences
```

---

# 106. Frontend Server State

Use:

```text
TanStack Query
```

for server state.

Do not create a large global store that duplicates the backend unless there is a real client-state requirement.

---

# 107. Frontend Form State

Use:

```text
React Hook Form
+
Zod
```

for form behavior.

Backend remains authoritative.

---

# 108. Frontend Financial Calculation Rule

Forbidden:

```text
client recalculates authoritative budget state

client calculates authoritative forecast

client calculates scenario result
```

Allowed:

```text
visual formatting

progress bar width from backend percentage

pure presentation transformations
```

---

# 109. Frontend Auth Bootstrap

Authentication state must include a loading/unknown state.

Do not cause:

```text
protected-route flicker
```

by assuming unauthenticated before `/auth/me` resolves.

---

# 110. Axios Rule

Use one central Axios client.

It owns:

```text
base URL

withCredentials

CSRF header

response interceptors
```

Do not create separate uncoordinated Axios instances unless intentionally needed.

---

# 111. 401 Rule

Frontend must use:

```text
single-flight refresh
```

Three simultaneous 401 responses must not generate three independent refresh requests.

---

# 112. Refresh Retry Rule

Each failed original request may retry at most once after successful refresh.

Never create an infinite refresh loop.

---

# 113. 403 Rule

Do not automatically log out on:

```text
403
```

A user may be authenticated but unauthorized.

---

# 114. 422 Rule

Expose field validation errors to forms.

Do not reduce all validation failures to a generic toast.

---

# 115. 409 Rule

Version conflict should produce dedicated UI.

Example:

```text
This record changed since you opened it.

Reload the latest version.
```

---

# 116. 429 Rule

Do not automatically spam retries.

Respect:

```text
Retry-After
```

when available.

---

# 117. Logout Rule

On logout or failed refresh:

```text
clear authentication state

clear user-sensitive TanStack Query cache
```

This prevents data leakage between users in the same browser.

---

# 118. Browser Storage Rule

Do not persist sensitive financial server responses to browser storage unless explicitly required.

No auth credentials in:

```text
localStorage

sessionStorage
```

---

# 119. UI Design Rule

Savio must not look like:

```text
a generic admin template
```

The interface should feel:

```text
calm

clear

trustworthy

financially intelligent

forward-looking
```

---

# 120. UI Information Hierarchy

For intelligent financial screens:

```text
1. Deterministic Result
2. Supporting Facts
3. Assumptions
4. AI Explanation
5. User Action
```

Do not place AI narrative above authoritative results.

---

# 121. Financial State Labels

The UI must clearly distinguish:

```text
ACTUAL

PROJECTED

SCENARIO

AI-GENERATED
```

If those become visually ambiguous, implementation is incorrect.

---

# 122. Loading States

Every major screen needs:

```text
loading
```

Prefer skeletons for predictable layouts.

Do not use a full-page spinner for every screen.

---

# 123. Empty States

Every list/intelligence screen needs an actionable empty state.

Bad:

```text
No data.
```

Preferred:

```text
No transactions yet.

Add your first income or expense to start understanding your cashflow.
```

---

# 124. Error States

Errors should answer:

```text
what failed

whether user data remains safe

whether retry is possible
```

---

# 125. Responsive Rule

Critical pages must be reviewed at:

```text
375px

768px

1024px

1440px
```

Do not treat "no horizontal overflow" as sufficient responsive design.

---

# 126. Table Rule

Large desktop tables should usually become:

```text
compact lists/cards
```

on mobile rather than compressed multi-column tables.

---

# 127. Accessibility Rule

At minimum support:

```text
keyboard navigation

visible focus

proper labels

semantic HTML

dialog focus management

status not communicated by color alone
```

---

# 128. AI UI Rule

AI content should use:

```text
subtle AI identity
```

not distracting glowing interfaces.

Avoid:

```text
unnecessary sparkles

fake thinking animations

AI-first dashboard hierarchy
```

---

# 129. Copy Rule

Use clear user-facing language.

Prefer:

```text
Add Expense
```

over:

```text
Create Expense Transaction Record
```

Prefer:

```text
Projected Balance
```

over:

```text
Prediction
```

if certainty is not justified.

---

# 130. Testing Rule

Do not consider a feature complete solely because it works manually.

Every P0 feature must include relevant automated tests.

---

# 131. Financial Feature Tests

For financial features consider:

```text
happy path

validation

authorization

financial invariant

rollback

concurrency

voiding

stale derived state
```

as applicable.

---

# 132. Security Tests

Required high-value tests include:

```text
CSRF

refresh rotation

refresh replay

cross-workspace IDOR

role enforcement

rate limiting

query validation
```

---

# 133. AI Tests

Use mock provider.

Test:

```text
valid output

invalid output

timeout

provider unavailable

AI disabled

cross-workspace tool access

prompt injection

unknown action
```

---

# 134. Frontend Tests

High-value:

```text
auth bootstrap

single-flight refresh

422 form mapping

409 conflict

AI degraded UI

scenario comparison

responsive critical interactions
```

---

# 135. Concurrency Tests

Explicitly test:

```text
refresh rotation race

recurring occurrence duplicate processing

optimistic version conflict
```

and any financial operation that uses mutable cached state.

---

# 136. Go Race Detector

Before submission run:

```bash
go test -race ./...
```

A clean race detector does not replace database concurrency tests.

---

# 137. CI Rule

A completed task must not knowingly break:

```text
backend tests

frontend tests

typecheck

build

migrations
```

---

# 138. Documentation Rule

Implementation changes that alter product behavior must update docs.

Do not knowingly leave documentation incorrect.

Examples requiring doc changes:

```text
new endpoint

changed enum

changed status lifecycle

new database relation

changed auth behavior

changed forecast assumption
```

---

# 139. OpenAPI Rule

If endpoint implementation changes:

```text
OpenAPI changes in the same milestone
```

Do not defer API documentation until the end for core endpoints.

---

# 140. Coding Agent Task Scope

When given a specific task:

```text
implement only the requested scope
```

Do not opportunistically refactor unrelated modules unless required.

---

# 141. No Scope Creep

Before adding anything not requested, ask:

```text
Is this required for the current milestone?

Does the source of truth require it?

Is it necessary for correctness/security?
```

If no:

```text
do not add it.
```

---

# 142. No Premature Bonus Features

Do not implement:

```text
Kafka

Kubernetes

GraphQL

Elasticsearch

microservices

vector database

dark mode

receipt OCR

advanced model routing
```

unless explicitly requested after P0 is stable.

---

# 143. No Premature Cache

Do not add Redis cache to ordinary financial reads just because Redis exists.

Measure first.

Redis initial purposes:

```text
queue

rate limiting

short-lived coordination
```

---

# 144. No Premature Event Bus

Savio does not need a full event-driven architecture for P0.

Use:

```text
service calls
+
background queue
```

unless an explicit reliability requirement justifies more.

---

# 145. No Premature CQRS

Separate read queries where useful.

Do not introduce a formal CQRS framework.

---

# 146. No Event Sourcing

Do not convert the project into full event sourcing.

Financial history already uses explicit domain records.

---

# 147. No Vector Database

AI presence does not automatically justify embeddings/vector search.

Do not introduce:

```text
pgvector
```

unless a concrete future semantic-search feature requires it.

---

# 148. No Autonomous AI Agent

Savio Copilot is:

```text
bounded orchestration
```

not an unrestricted autonomous agent.

Do not give it broad execution capability.

---

# 149. Performance Rule

Optimize only after identifying a real issue.

Initial priorities:

```text
correct indexes

bounded pagination

no N+1

database aggregation

request context cancellation
```

---

# 150. N+1 Rule

When listing related resources:

```text
avoid one query per row
```

Use:

```text
joins

preload

batch lookup
```

appropriately.

---

# 151. Search Rule

Search should be:

```text
workspace-scoped

parameterized

bounded
```

P0 may use PostgreSQL:

```text
ILIKE
```

No need for Elasticsearch.

---

# 152. Pagination Rule

Default:

```text
page = 1

limit = 20
```

Maximum:

```text
100
```

unless API contract says otherwise.

---

# 153. Sorting Rule

Allowed sort fields are explicit.

Unknown sort:

```text
validation error
```

---

# 154. Health Rule

`/health` answers:

```text
is process alive?
```

It should not fail merely because AI provider is unavailable.

---

# 155. Readiness Rule

Critical:

```text
PostgreSQL
```

AI is degradable.

Redis criticality depends on current feature path, but should be reported clearly.

---

# 156. Logging Rule

Use structured logs.

Useful fields:

```text
request_id

user_id

workspace_id

method

path

status

duration

error_code
```

---

# 157. Sensitive Logging Rule

Never log:

```text
password

access token

refresh token

CSRF token

AI API key

database credential

full sensitive financial context by default
```

---

# 158. Request ID Rule

Every HTTP request should have a request ID.

Propagate to:

```text
logs

AI request metadata

background jobs where useful
```

---

# 159. Audit Rule

Important business/security events should be auditable.

Audit logs must never contain secrets.

---

# 160. Object Storage Rule

When file features are implemented:

```text
binary file
→ MinIO

metadata
→ PostgreSQL
```

Do not use local container filesystem as persistent storage.

---

# 161. Upload Security Rule

If file upload is implemented:

```text
authenticate

authorize

limit size

allowlist MIME

generate object path

keep object private
```

---

# 162. Git Rule

Make focused, meaningful commits.

Preferred:

```text
feat(auth): implement refresh token rotation
```

Avoid:

```text
update
fix stuff
final
```

---

# 163. Generated File Rule

Do not modify generated files manually unless the generation strategy requires it.

If OpenAPI/client generation exists:

```text
modify source

regenerate
```

---

# 164. Formatting

Go:

```bash
gofmt
```

Frontend:

```text
project lint/formatter conventions
```

Do not submit unformatted code.

---

# 165. Build Verification

Before declaring completion:

Backend:

```bash
go test ./...
go build ./...
```

Frontend:

```bash
npm run typecheck
npm run test
npm run build
```

Use actual repository commands if they differ.

---

# 166. Migration Verification

For schema changes verify:

```text
up

down

up
```

where practical.

At minimum:

```text
fresh DB migrate-up succeeds
```

---

# 167. Runtime Verification

Do not rely only on compiler/tests when implementing major integration behavior.

For high-value flows verify actual runtime:

```text
login

financial write

forecast

scenario

AI mock
```

as appropriate.

---

# 168. Agent Completion Report

When completing a coding task, report:

```text
What changed

Why

Files changed

Tests added

Commands run

Results

Known limitations

Documentation updated
```

Do not claim verification that was not actually performed.

---

# 169. Agent Honesty Rule

Never claim:

```text
all tests pass
```

without running them.

Never claim:

```text
endpoint verified
```

without testing it.

Never claim:

```text
no regression
```

without evidence.

---

# 170. Failure Reporting

If blocked:

```text
state the exact blocker

state what was completed

state what remains

state the smallest next action
```

Do not hide incomplete work behind a success summary.

---

# 171. Existing Code Rule

Before rewriting existing code:

```text
read it

understand dependencies

identify why it is insufficient
```

Avoid large rewrites solely for stylistic preference.

---

# 172. Refactoring Rule

Refactor when:

```text
required for correctness

required for testability

required for architecture boundary

substantial duplication causes risk
```

Do not refactor unrelated modules during a focused feature task.

---

# 173. Database Schema Change Rule

Never silently change existing semantics.

Before migration ask:

```text
Does this alter authority?

Does it alter financial calculations?

Does it invalidate existing data?

Does API need to change?
```

---

# 174. Backward Compatibility

During P0 development, strict backward API compatibility is less important than consistency.

However, when changing a contract:

```text
update all consumers immediately
```

Do not leave mixed old/new behavior.

---

# 175. Enum Consistency Rule

Enums are contracts.

A change such as:

```text
CANCELLED
→ ENDED
```

requires updates to:

```text
database constraints

Go constants

API docs

frontend types

UI labels

tests
```

---

# 176. Status Consistency Rule

Do not use multiple words for the same lifecycle state across modules unless semantically distinct.

Keep lifecycle vocabulary deliberate.

---

# 177. Financial Formula Versioning

Forecast/scenario algorithms may include:

```text
calculation_version
```

When materially changing algorithm behavior, increment the version if persisted snapshots depend on it.

---

# 178. Clock Rule

Time-dependent domain logic should use an injectable clock abstraction.

Avoid scattered direct calls to:

```go
time.Now()
```

inside deterministic domain calculations.

---

# 179. Context Cancellation

Propagate request context through:

```text
handler

service

repository

AI provider
```

Do not replace request context with `context.Background()` casually.

---

# 180. External Call Timeouts

Every external call needs a timeout.

Especially:

```text
AI provider

MinIO where relevant
```

No unlimited external network requests.

---

# 181. Retry Rule

Retry only:

```text
bounded

transient

idempotent/safe
```

operations.

Do not retry arbitrary financial writes automatically without an idempotency model.

---

# 182. Idempotency

Consider idempotency for:

```text
transfers

recurring occurrence confirmation

background jobs
```

If public transaction creation uses `Idempotency-Key`, follow the API contract exactly.

---

# 183. Security Before Convenience

Do not weaken:

```text
CSRF

authorization

cookie security

validation

AI tool restrictions
```

because they complicate implementation.

Solve the implementation correctly.

---

# 184. UI Before Bonus

After backend correctness is stable, prioritize:

```text
polished core UI
```

before bonus integrations.

Assessment gives significant weight to UI/UX.

---

# 185. Critical Screens

Highest frontend priority:

```text
Login

Onboarding

Dashboard

Transactions

Forecast

Scenario Simulator

AI Insights

Savio Copilot
```

---

# 186. Demo Coherence Rule

The final project should demonstrate one connected story.

Avoid building features that never connect to the main user journey.

Desired demo:

```text
create finance data
↓
dashboard understands it
↓
forecast projects it
↓
scenario modifies it hypothetically
↓
AI explains it
```

---

# 187. Product Language Rule

Use user-facing terminology from product docs.

Examples:

```text
Cashflow Forecast

Scenario Simulator

AI Insight

Savio Copilot
```

Avoid internal engineering names in UI.

---

# 188. AI Naming Rule

Do not label deterministic outputs as:

```text
AI Forecast
```

Forecast belongs to Finance Engine.

AI can provide:

```text
Forecast Explanation
```

---

# 189. Scenario Naming Rule

Do not call scenario result:

```text
AI Prediction
```

It is a:

```text
deterministic scenario projection
```

---

# 190. Actual vs Projected Rule

Always distinguish:

```text
actual transactions

planned recurring occurrences

forecast estimates

scenario modifications
```

Do not mix them into one indistinguishable ledger list.

---

# 191. Data Deletion Rule

Financial-history deletion must be deliberate.

Prefer:

```text
archive

void
```

over hard delete once history exists.

---

# 192. Demo Data Rule

Use synthetic realistic data.

Do not seed real personal data.

---

# 193. Testing Data Rule

Tests must create their own isolated data.

Do not rely on demo seed for automated correctness.

---

# 194. Code Comment Rule

Comments should explain:

```text
why
```

not restate obvious code.

Useful:

```text
Lock membership row to prevent two concurrent owner demotions from removing the last owner.
```

Not useful:

```text
// increment i
i++
```

---

# 195. TODO Rule

Do not leave unresolved P0 behavior behind casual `TODO`.

If something is intentionally deferred:

```text
document it as P1/P2
```

and ensure current behavior is complete.

---

# 196. Panic Rule

Do not use `panic` for ordinary business errors.

Return typed errors.

Panic is for truly unrecoverable programmer/startup conditions.

---

# 197. Error Wrapping

Preserve internal cause for logging while returning safe application errors.

Do not discard useful context.

---

# 198. Domain Error Codes

Use stable codes such as:

```text
VALIDATION_ERROR

RESOURCE_NOT_FOUND

VERSION_CONFLICT

PERMISSION_DENIED

CSRF_TOKEN_INVALID

AI_PROVIDER_UNAVAILABLE
```

Do not invent new error strings inconsistently when an existing semantic code fits.

---

# 199. API DTO Rule

Do not serialize GORM models directly as public API response when it exposes:

```text
internal fields

foreign keys

timestamps

sensitive metadata
```

Use explicit response DTOs where useful.

---

# 200. Sensitive Model Fields

Never accidentally serialize:

```text
password_hash

refresh_token_hash

internal AI prompt

secret metadata
```

---

# 201. Database Model vs Domain Model

They may be the same for simple resources, but do not force them to be identical when domain semantics require separation.

---

# 202. System Category Rule

Seed system categories through an idempotent seed process.

Do not create duplicate system categories on each startup.

---

# 203. Bootstrap Rule

Application startup must not mutate production schema automatically.

Migrations are an explicit deployment/developer operation.

---

# 204. Seed Rule

Do not automatically inject demo financial data into normal production startup.

Development seed must be explicit.

---

# 205. Configuration Rule

Validate configuration at startup.

Missing critical configuration:

```text
fail fast
```

Optional AI configuration may allow degraded startup when AI is disabled.

---

# 206. AI Startup Rule

If:

```text
AI_ENABLED=false
```

the application should not require:

```text
AI_API_KEY
```

---

# 207. Readiness Rule for AI

AI outage should normally report:

```text
degraded
```

not make financial API unready.

---

# 208. Redis Failure Policy

When Redis is unavailable:

```text
do not silently change financial correctness
```

Async features may degrade.

Security-sensitive rate-limit behavior must follow documented fail-open/fail-closed policy.

---

# 209. MinIO Failure Policy

If no upload action is being performed:

```text
finance features remain functional
```

---

# 210. Dependency Addition Rule

Before adding a new dependency, ask:

```text
What problem does it solve?

Can existing stack solve it cleanly?

Is the dependency maintained?

Does it create unnecessary complexity?
```

---

# 211. Frontend Dependency Rule

Avoid UI libraries that conflict heavily with the custom Tailwind design unless explicitly selected.

A small headless component library may be acceptable.

Do not introduce multiple overlapping component systems.

---

# 212. Accessibility Before Animation

If time is constrained:

```text
accessibility
```

wins over:

```text
animations
```

---

# 213. Correctness Before Caching

If time is constrained:

```text
correct deterministic calculations
```

wins over:

```text
Redis cache
```

---

# 214. Tests Before Bonus

If time is constrained:

```text
critical tests
```

win over:

```text
additional AI features
```

---

# 215. Documentation Before Hidden Complexity

If a behavior is too complex to explain clearly in docs:

```text
simplify the behavior
```

unless the complexity is genuinely required.

---

# 216. PR Review Checklist

Before marking any task done, inspect:

```text
Architecture

Business Rules

Security

Database Integrity

API Contract

Testing

Frontend State

Documentation
```

---

# 217. Backend Review Checklist

```text
[ ] Handler is thin

[ ] Service owns business logic

[ ] Resource is workspace-scoped

[ ] Authorization is backend enforced

[ ] Inputs validated

[ ] Money uses decimal-safe types

[ ] Transaction boundary correct

[ ] Concurrency considered

[ ] Errors standardized

[ ] Tests added
```

---

# 218. Database Review Checklist

```text
[ ] Migration up exists

[ ] Migration down exists where practical

[ ] Foreign keys correct

[ ] Unique constraints correct

[ ] Indexes intentional

[ ] No derived field incorrectly treated as authority

[ ] Financial precision safe
```

---

# 219. API Review Checklist

```text
[ ] Route under /api/v1

[ ] Correct HTTP method

[ ] Correct status codes

[ ] Success envelope consistent

[ ] Error envelope consistent

[ ] Pagination bounded

[ ] Sort allowlisted

[ ] OpenAPI updated
```

---

# 220. Security Review Checklist

```text
[ ] Authentication required where needed

[ ] CSRF required for mutation

[ ] Role permission checked

[ ] Workspace scope checked

[ ] No IDOR

[ ] Rate limit considered

[ ] Sensitive data not logged

[ ] Secrets server-only
```

---

# 221. Financial Review Checklist

```text
[ ] Actual vs projected semantics clear

[ ] Transfers excluded from income/expense

[ ] VOIDED records excluded

[ ] Adjustments treated correctly

[ ] Balance reconstructable

[ ] Forecast deterministic

[ ] Scenario non-destructive
```

---

# 222. AI Review Checklist

```text
[ ] Does AI need to be involved?

[ ] Deterministic facts calculated first

[ ] Context minimized

[ ] Authorization context injected by backend

[ ] Tools bounded

[ ] Output schema validated

[ ] AI failure isolated

[ ] User remains in control
```

---

# 223. Frontend Review Checklist

```text
[ ] Loading state

[ ] Empty state

[ ] Error state

[ ] Pending state

[ ] Form validation

[ ] Server validation mapping

[ ] Responsive

[ ] Accessible

[ ] Financial truth comes from backend

[ ] AI visually differentiated
```

---

# 224. Test Review Checklist

```text
[ ] Happy path

[ ] Validation failure

[ ] Authorization failure

[ ] Cross-workspace case

[ ] Business-rule edge

[ ] Concurrency if relevant

[ ] Regression case if fixing bug
```

---

# 225. Documentation Review Checklist

```text
[ ] API docs still accurate

[ ] Database docs still accurate

[ ] Business rules still accurate

[ ] Enum names consistent

[ ] README commands still work
```

---

# 226. Required Verification Before Completion

At minimum run the tests relevant to changed code.

For broad changes, run:

```bash
go test ./...
go test -race ./...
```

Frontend:

```bash
npm run typecheck
npm run test
npm run build
```

Migration changes:

```text
fresh migrate-up
```

---

# 227. Do Not Fabricate Verification

If a command cannot be run:

```text
say so explicitly
```

Do not substitute:

```text
"It should work."
```

for actual verification.

---

# 228. Handling Existing Test Failures

If pre-existing unrelated tests are failing:

```text
identify them explicitly

show they are pre-existing if verified

do not hide them
```

Do not modify unrelated tests merely to achieve green output.

---

# 229. Fixing Bugs

When fixing a bug:

```text
reproduce

add failing regression test

fix

run test

review related invariants
```

---

# 230. Security Bug Rule

Every confirmed security bug requires a regression test.

---

# 231. Financial Bug Rule

Every confirmed financial correctness bug requires a regression test.

---

# 232. Critical Invariants

The following invariants must never be violated:

```text
INV-001
A user cannot access a workspace without active membership.

INV-002
VIEWER cannot mutate financial state.

INV-003
The last OWNER cannot be removed or demoted.

INV-004
Money arithmetic is decimal-safe.

INV-005
Account balance is reconstructable from authoritative financial history.

INV-006
Transfers do not change total internal portfolio value.

INV-007
Transfers do not count as income or expense.

INV-008
VOIDED transactions do not count as active income or expense.

INV-009
Reconciliation creates an adjustment rather than rewriting history.

INV-010
A recurring occurrence becomes actual at most once.

INV-011
Forecast output is deterministic from its inputs and assumptions.

INV-012
Scenario calculation does not mutate actual finance state.

INV-013
AI never chooses trusted authorization context.

INV-014
AI never becomes source of authoritative financial calculations.

INV-015
AI writes require deterministic validation and explicit user confirmation.

INV-016
Authentication credentials are never stored in localStorage/sessionStorage.

INV-017
Cookie-authenticated mutations are CSRF protected.

INV-018
Refresh token rotation invalidates the previous token.

INV-019
Cross-workspace resource IDs never expose data.

INV-020
Financial writes are validated server-side.
```

---

# 233. Invariant Priority

When implementation convenience conflicts with an invariant:

```text
the invariant wins
```

Do not weaken an invariant without explicit architecture change.

---

# 234. Core Demo Flow

Agents should preserve the ability to demo:

```text
Register
↓
Default Workspace
↓
Create Account
↓
Add Income
↓
Add Expense
↓
Create Budget
↓
View Dashboard
↓
Calculate Forecast
↓
Create Scenario
↓
Compare Baseline vs Scenario
↓
AI Explanation
↓
Ask Copilot
↓
Review Supporting Facts
```

---

# 235. RBAC Demo Flow

Preserve:

```text
OWNER adds VIEWER

VIEWER can read

VIEWER mutation → 403

OWNER changes VIEWER → MEMBER

MEMBER can mutate
```

This is important for assessment coverage.

---

# 236. AI Degraded Demo

The application must support:

```text
AI unavailable
```

while:

```text
transactions

analytics

budgets

forecast

scenario
```

remain functional.

---

# 237. Final Engineering Hierarchy

All code should preserve this hierarchy:

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
FORECAST / SCENARIO
    ↓
AI INTERPRETATION
    ↓
FRONTEND PRESENTATION
    ↓
USER DECISION
```

---

# 238. Final Agent Rule

Before writing code, ask:

```text
What is the authoritative source?

Which business rule applies?

Which security boundary applies?

Which financial invariant applies?

Does this require AI at all?

What tests prove it?

Which documentation must remain aligned?
```

Before declaring the task complete, ask:

```text
Did I implement only the intended scope?

Did I preserve financial correctness?

Did I preserve authorization?

Did I verify the implementation?

Did I update the source of truth?
```

The final Savio principle remains:

> **Finance Engine calculates. AI interprets. User decides.**
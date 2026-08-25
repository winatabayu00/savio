# Savio — Interview Preparation

## Project One-Liner

Savio is a personal cashflow intelligence platform where deterministic finance services calculate actuals, forecasts, and scenarios, while bounded AI explains the results.

## 60-Second Pitch

I built Savio for people who need more than a history of expenses. It connects tracking, cashflow understanding, deterministic forecasting, non-destructive scenario simulation, and grounded explanations. The backend is a Go/Gin/GORM modular monolith over PostgreSQL. Financial records are workspace-scoped and authoritative; account balances, analytics, forecasts, and scenarios are calculated by backend services using integer minor units. Authentication uses HttpOnly cookies, CSRF protection, rotating refresh sessions, and backend-enforced OWNER/MEMBER/VIEWER permissions. The React frontend uses feature modules, TanStack Query, React Hook Form, Zod, Tailwind, and one Axios client. AI is deliberately downstream: it receives bounded, authorized facts and interprets them. It cannot calculate authoritative balances or mutate financial state.

## 3-Minute Pitch

Savio addresses a common gap in expense trackers: users can see what happened but struggle to understand why cashflow changed, what may happen next, or whether a decision is safe to test. The product flow is Track → Understand → Forecast → Simulate → Explain → Decide.

The backend is a modular monolith. Gin handlers bind and map requests, application services own authorization and business workflows, repositories own PostgreSQL queries, and platform packages provide money, errors, auth context, rate limiting, and middleware. PostgreSQL remains the financial source of truth. Money uses integer minor units; posted transactions are immutable; corrections use void plus replacement. Transfers are separate from income and expense and must be atomic. Recurring activity separates planned occurrences from actual posted transactions.

Forecast and scenario calculations are deterministic. Forecast combines current state, known events, scheduled recurring activity, estimated variable spending, and explicit assumptions. A scenario overlays changes on a baseline and never mutates actual records. Source revisions make stale results visible.

Security uses cookie authentication, Argon2 password hashing, signed CSRF tokens, server-side refresh sessions, workspace membership checks, and OWNER/MEMBER/VIEWER permissions. The frontend has an auth bootstrap unknown state and a single-flight Axios refresh path. AI is bounded behind an orchestrator and finance tools; it interprets structured facts from authorized services rather than querying PostgreSQL or deciding trusted scope.

The honest trade-off is P0 simplicity: one base currency per workspace, application-layer tenancy instead of PostgreSQL RLS, a modular monolith instead of microservices, and limited frontend/E2E coverage. The current submission also needs refresh-rotation race hardening and green integration tests.

## Product Defense

### Basic: Why Savio?

**Short Answer:** Expense tracking explains the past. Savio adds explainable planning before a decision.

**Deep Answer:** The product joins authoritative records to deterministic analytics, forecast, and scenario workflows. AI is an explanation layer, not a financial authority.

**Implementation Evidence:** `README.md`; `backend/internal/forecast/`; `backend/internal/scenarios/`; `backend/internal/ai/`; frontend forecast/scenario/Copilot pages.

**Trade-Off:** Broader product scope increases UI/test surface. It was accepted because every feature supports one connected cashflow journey.

**Likely Follow-Up:** How do you prove it is not CRUD?

**Follow-Up Answer:** The system enforces lifecycle, transfer, recurring, authorization, freshness, and non-destructive scenario rules. Those rules change outcomes; CRUD alone does not.

### Intermediate: Who is the target user?

**Answer:** Individuals who need day-to-day cashflow visibility and safer planning, not traders or investment users. P0 intentionally avoids FX conversion, investment portfolios, and autonomous financial action.

### Deep Dive: What is the product authority hierarchy?

**Answer:** PostgreSQL financial records are authoritative. Domain services calculate balances, analytics, forecasts, and scenarios. AI receives validated facts and interprets them. The user remains the decision-maker.

## Architecture Defense

### Basic: Why a modular monolith?

**Short Answer:** It preserves clear boundaries and database transactions without premature distributed complexity.

**Deep Answer:** Savio needs atomic financial writes, shared authorization, simple local setup, and reviewable code. Modules provide separation while one process and PostgreSQL keep transfer, reconciliation, and lifecycle transactions straightforward.

**Evidence:** `backend/internal/*`; `docs/architecture/system-architecture.md`.

**Trade-Off:** Less independent scaling and isolation than microservices. Revisit when module ownership, queue throughput, or deployment isolation becomes a measured bottleneck.

**Follow-Up:** When split services?

**Answer:** Only after identifying a real boundary: for example AI workloads needing isolated scaling or analytics reads requiring independent capacity. Financial writes should remain transactionally close until a deliberate distributed-consistency design exists.

### Basic: Why Go + Gin + GORM?

**Answer:** Go gives a small, explicit runtime and strong concurrency primitives; Gin is lightweight HTTP routing; GORM reduces repetitive persistence mapping while repositories keep queries behind module boundaries.

### Challenge: Why not put business logic in repositories?

**Answer:** Repositories know persistence, not policy. Services own authorization, state transitions, transaction boundaries, and cross-module workflows. This keeps rules testable and prevents query code from becoming the domain authority.

### Architecture Whiteboard — 60 Seconds

```text
React feature pages
        ↓ Axios / cookies / CSRF
Go Gin API
        ↓
Handlers → Application services → Domain/finance rules → Repositories
                                                    ↓
                                                PostgreSQL

API/services → Redis → Go worker
AI orchestrator → bounded finance tools → deterministic services → structured facts → AI provider
```

Say: “PostgreSQL is authoritative. Redis is async/coordination infrastructure. AI is downstream of deterministic, authorized facts.”

### Architecture Whiteboard — 3 Minutes

Explain request validation and auth first, then service-owned authorization and transaction boundaries, then repository persistence. Draw the worker as a caller of shared services, not a duplicate domain implementation. Draw AI below the finance services, never beside them as an authority. Mention current limitation: Docker Compose currently packages infrastructure, not the full API/frontend/worker deployment.

## Database Q&A

### Basic: Why PostgreSQL?

**Answer:** Financial relationships, constraints, transactional writes, indexed filtering, and aggregations are central. PostgreSQL handles those directly and gives a defensible relational model.

### Basic: Why explicit migrations instead of AutoMigrate?

**Answer:** Financial schema changes need reviewable, reproducible, reversible history. Explicit golang-migrate up/down files make zero-to-current setup and rollback visible; GORM AutoMigrate is not the production schema authority.

**Evidence:** `backend/internal/migrations/`, `migrations_test.go`.

### Basic: Why BIGINT minor units?

**Answer:** Integer minor units avoid floating-point money errors. API values remain decimal-safe strings and are parsed at the boundary.

### Challenge: Why do analytics use `float64` sometimes?

**Answer:** Percentages and presentation ratios are metadata, not authoritative money. Stored/calculated monetary amounts remain integer-based. I would still keep the distinction explicit in DTOs and tests.

### Basic: How is balance calculated?

**Answer:** Opening balance plus posted ledger effects equals derived account balance. There is no independently mutable current balance as authority.

### Deep Dive: Why not store `current_balance`?

**Answer:** A mutable balance can drift from history. Derived balance is reconstructable and auditable. At scale I could add a materialized projection, but it would remain rebuildable derived state protected by transactional updates.

### Challenge: Why not double-entry accounting?

**Answer:** P0 models personal cashflow with accounts, transactions, and separate transfers. Full double-entry would improve accounting rigor but adds ledger-leg complexity beyond the assignment’s target. If Savio handled liabilities, FX, investments, or formal reconciliation at scale, I would revisit it.

### Database Whiteboard

```text
Workspace
  ├── Account
  ├── Transaction
  ├── Transfer (source + destination)
  ├── Recurring Rule → Occurrence → Posted Transaction
  ├── Budget / Goal
  ├── Forecast Run (source revision)
  └── Scenario → modifications → Scenario Run
```

Mention foreign keys, workspace-scoped queries, unique membership/occurrence constraints, positive amount checks, and intentional date/workspace indexes.

## Financial Domain Q&A

### Basic: How do posted corrections work?

**Answer:** Posted financial fields are immutable. The original is voided, then a replacement is created. History remains visible and active analytics exclude the voided record.

### Basic: How does transfer posting work?

**Answer:** A transfer validates distinct active same-workspace accounts and positive amount, then writes both sides atomically. It changes account distribution, not total portfolio value, and is excluded from income/expense analytics.

### Deep Dive: How do you prevent partial transfers?

**Answer:** The transfer service uses one database transaction for source and destination effects. A failed operation rolls back both. The important invariant is no successful state where only one side changed.

### Basic: How does recurring confirmation work?

**Answer:** Recurring rules represent planned activity. Occurrences begin pending and become actual only once confirmed or explicitly auto-posted. Database uniqueness and service checks prevent duplicate actualization.

### Challenge: What if two workers confirm one occurrence?

**Answer:** The operation needs a transaction plus a unique database constraint/state transition so only one worker wins. The existing code has recurring idempotency tests; I would show the exact constraint and row-lock behavior rather than rely on Redis or worker memory.

### Basic: How do budgets work?

**Answer:** A monthly category budget derives actual spend from posted expense transactions matching category and period. Transfers, voided transactions, and ordinary adjustments are excluded.

### Basic: Why is forecast deterministic?

**Answer:** Users need reproducible, explainable financial projections. Forecast combines current state, known events, recurring schedules, historical estimates, and explicit assumptions. AI can explain the result but must not calculate it.

### Deep Dive: How is forecast freshness determined?

**Answer:** Forecast output records source financial revision/freshness metadata. Relevant finance changes invalidate the snapshot, so the UI can distinguish current from stale projections.

### Challenge: What if there is limited history?

**Answer:** Reduce or qualify estimated-variable-spend confidence, expose assumptions, and rely more on known/scheduled events. Do not invent precision. The forecast remains deterministic from available inputs.

### Basic: How does Scenario Simulator work?

**Answer:** Baseline forecast plus hypothetical modifications equals scenario projection. Scenario calculation persists scenario state/snapshot only; it does not mutate accounts, transactions, recurring rules, or budgets.

### Challenge: What if finance data changes after scenario calculation?

**Answer:** Source revision changes mark the scenario stale. The user must recalculate rather than silently treating old output as current.

## Authentication Q&A

### Basic: Why cookie auth instead of localStorage?

**Answer:** HttpOnly cookies keep credentials out of JavaScript and reduce token theft exposure through XSS. Since cookies are sent automatically, CSRF protection is required separately.

### Basic: What do cookie attributes do?

**Answer:** `HttpOnly` blocks JavaScript token reads; `Secure` limits transport to HTTPS in production; `SameSite` reduces cross-site sending; narrow `Path` limits exposure; `Max-Age`/`Expires` bounds lifetime.

### Basic: Why store refresh tokens hashed?

**Answer:** The browser holds the raw opaque token; PostgreSQL stores only a hash. A database read should not become an immediately usable session credential.

### Authentication Whiteboard

```text
Login
  ↓ access cookie + refresh cookie
Access expires
  ↓ API returns 401
Axios single-flight refresh
  ↓
Server validates session, rotates refresh token, issues new cookies
  ↓
Retry original request once
  ↓ failure: clear auth/query cache and redirect
```

### Deep Dive: What happens when two refreshes arrive concurrently?

**Accurate answer for current code:** The frontend shares one refresh promise, which prevents normal browser duplication. The backend intends rotation tests, but the implementation currently looks up the old hash before locking and does not re-check it inside the locked transaction. Therefore I would not claim the backend race is solved. The fix is lock-first, compare-inside-transaction, conditional rotation, and a regression test that reliably fails before the fix.

### Challenge: How do you revoke a session?

**Answer:** Refresh sessions are server-tracked and can be revoked; logout revokes the session and clears cookies. Access tokens are short-lived. Be precise about implemented individual/session-family revocation rather than claiming a broader session-management UI.

### Challenge: What if the user is disabled?

**Answer:** Authentication should reject new and existing use. The current refresh path needs an explicit active-user recheck; I would fix that before claiming disabled-user session invalidation is complete.

## CSRF Q&A

### Basic: CSRF vs XSS?

**Answer:** CSRF tricks a browser into sending its ambient cookies from another site. XSS executes attacker JavaScript in the trusted origin. HttpOnly helps against JavaScript reading cookies, but it does not stop a cross-site request carrying cookies, so CSRF protection remains necessary.

### Deep Dive: How does Savio CSRF work?

**Answer:** The backend exposes a CSRF token endpoint. The frontend fetches it once with single-flight coordination and sends it in `X-CSRF-Token` on POST/PUT/PATCH/DELETE. Middleware validates the signed token for every mutation. SameSite is defense in depth, not the sole control.

### Challenge: What if an attacker guesses a token?

**Answer:** The token is signed with a server secret and must match the expected structure/signature. The secret must be strong and required at startup; current permissive secret validation is a defect to fix.

## RBAC & IDOR Q&A

### Basic: How do OWNER/MEMBER/VIEWER work?

**Answer:** OWNER manages workspace and members, MEMBER reads/writes finance, VIEWER is read-only. The backend checks active membership and role for every protected operation.

### Basic: Why do UUIDs not prevent IDOR?

**Answer:** UUID unpredictability is not authorization. Every repository lookup still needs authenticated workspace scope. Savio uses workspace-scoped queries rather than `WHERE id = ?` alone.

### Challenge: What if a VIEWER manually calls POST?

**Answer:** Backend middleware/service permission checks return 403. Frontend hiding a button is only UX and is not trusted.

### Challenge: What if the AI supplies another `goal_id`?

**Answer:** AI IDs are untrusted. The normal service lookup still verifies workspace scope and permission. AI never chooses trusted user/workspace authorization context.

## Concurrency Q&A

### Basic: Where are transaction boundaries?

**Answer:** Financial operations that must be atomic, including transfer posting and relevant recurring/reconciliation workflows, use explicit database transactions in services. Refresh rotation also needs an explicit lock/conditional-update boundary.

### Intermediate: How does optimistic locking work?

**Answer:** Mutable resources carry a version; updates include the expected version. A stale update returns a conflict such as `VERSION_CONFLICT` rather than silently overwriting newer data.

### Challenge: Two users update the same budget. What happens?

**Answer:** One update succeeds; the stale version receives 409. The frontend should explain that the record changed and offer reload. I would verify the exact budget endpoint/test before claiming every mutable page has complete conflict UX.

### Challenge: Two transfers execute concurrently?

**Answer:** Each operation must validate and persist atomically. Database row locking/transaction isolation should protect account effects; total portfolio value must remain invariant. At higher scale I would test lock ordering to avoid deadlocks.

## Frontend Q&A

### Basic: Why TanStack Query?

**Answer:** Server state needs caching, invalidation, loading/error states, and query keys. TanStack Query handles that without duplicating authoritative finance data in a global store.

### Basic: Why React Hook Form + Zod?

**Answer:** They provide efficient form state and schema validation for fast UX. Backend validation remains authoritative; API field details should map back to form fields.

### Deep Dive: How does Axios refresh single-flight work?

**Answer:** `refreshPromise` is shared across simultaneous 401s. Each original request has `_retried`, so it retries at most once. Auth endpoints are excluded to prevent loops. Failure dispatches an unauthorized event; the auth provider clears auth/query state.

### Challenge: How do 403/422/429/500 behave?

**Accurate answer:** The client converts them into `ApiError`, but it does not currently provide the assignment’s complete centralized status-specific UX. Some pages handle errors locally. This is a gap, not something to overclaim.

### Challenge: How do workspace switches affect cache keys?

**Answer:** Query keys should include the active workspace or be invalidated on switch/logout. Sensitive cache must be cleared on logout. I would point to the actual provider/query-key implementation and avoid claiming a global cache policy not present in code.

### Challenge: Why not calculate budget totals in React?

**Answer:** Backend aggregation is authoritative and avoids disagreement caused by incomplete pages, timezone rules, voids, transfers, or stale client data. React formats and visualizes backend values.

## Forecast Q&A

### Basic: What inputs drive forecast?

**Answer:** Current derived liquid balance, future known activity, recurring schedules, historical variable spending where available, and explicit assumptions.

### Deep Dive: Why not ask an LLM to predict cashflow?

**Answer:** An LLM is probabilistic, difficult to reproduce, and weak as a financial authority. The deterministic engine exposes inputs, assumptions, event types, minimum balance, and calculation version. AI can produce a readable explanation afterward.

### Challenge: What does confidence mean?

**Answer:** It is deterministic metadata based on data quality and assumptions, not model confidence. Sparse history or many assumptions should reduce confidence and be visible to the user.

## Scenario Q&A

### Basic: What is a scenario?

**Answer:** A non-destructive overlay on a deterministic baseline forecast.

### Deep Dive: How do you guarantee no mutation?

**Answer:** Scenario service reads authoritative inputs and writes only scenario state/snapshot. It does not call transaction/account mutation paths. Tests should compare actual records before and after calculation.

### Challenge: Scenario after new expense?

**Answer:** Source revision changes make it stale. Recalculate from the new baseline; never silently merge old scenario output with current actuals.

## AI Q&A

### Basic: Why use AI at all?

**Answer:** AI adds natural-language interpretation, categorization suggestions, insight narratives, forecast explanations, and Copilot interaction. It does not replace deterministic analytics.

### Basic: How does AI access data?

**Answer:** Authenticated backend orchestration calls bounded allowlisted finance tools. Tools call domain services and return minimal structured facts. The model never gets arbitrary SQL, filesystem, shell, or trusted authorization context.

### Deep Dive: What prevents hallucinated financial numbers?

**Answer:** Authoritative numbers come from deterministic tool output. Structured model output is schema-validated before application use. The UI distinguishes AI-generated explanation from actual/projected/scenario values. The response should still be treated as interpretation, not a ledger fact.

### Deep Dive: What prevents prompt injection?

**Answer:** Transaction descriptions and notes are untrusted data. Tool capabilities are allowlisted by backend; model text cannot expand permissions or choose workspace identity. The model cannot execute arbitrary SQL or HTTP.

### Challenge: What if AI returns invalid JSON or an unknown tool?

**Answer:** Schema validation rejects invalid output and unknown actions. The endpoint returns a safe error/degraded state; no financial mutation occurs.

### Challenge: What if provider is unavailable?

**Answer:** AI endpoints degrade, while accounts, transactions, analytics, budgets, forecast, and scenario remain functional. Tests use a mock provider rather than live calls.

### AI Whiteboard

```text
User question
  ↓
Copilot / AI orchestrator
  ↓ authorized allowlisted tool
Finance service
  ↓
Structured facts
  ↓
LLM explanation
  ↓ schema validation
Safe AI response
```

## Security Q&A

### Challenge: Why need CSRF if SameSite is enabled?

**Answer:** SameSite is browser policy and deployment-dependent. It is defense in depth, not a complete application authorization proof. Explicit token validation protects state-changing endpoints when cookies are ambient.

### Challenge: Could AI read another user’s data?

**Answer:** It should not: user/workspace context is injected server-side and every tool resource lookup is workspace-scoped. AI-provided IDs are untrusted. This is tested through service authorization, not inferred from UUIDs.

### Challenge: Could I replay a refresh token?

**Accurate answer:** The intended design rejects rotated tokens, and tests cover replay intent. Actual code’s lock/hash sequencing is insufficient under concurrency, so replay/race hardening is a known submission blocker.

### Challenge: What protects dynamic sort SQL?

**Answer:** Sort fields map through an allowlist to known SQL expressions. Raw query sort text is rejected with validation error. The backend has injection-oriented sort tests.

### Challenge: How are secrets handled?

**Answer:** They come from environment configuration and are not committed. Current config logs weak/missing JWT/CSRF secrets instead of failing startup; that must be corrected before production claims.

## Testing Q&A

### Basic: What is your testing strategy?

**Answer:** Test financial invariants and security boundaries first, then frontend auth/interceptor behavior, then integration/runtime paths. The suite covers transaction lifecycle, transfers, recurring, budgets, forecast, scenarios, RBAC, IDOR, CSRF, refresh behavior, errors, and voice UI.

### Deep Dive: Which tests would impress a reviewer?

**Answer:** Last-owner protection, cross-workspace IDOR, VIEWER mutation denial, transfer invariants, voiding, recurring idempotency, CSRF mismatch, refresh replay/concurrency intent, and frontend single-flight refresh.

### Challenge: Is the suite green?

**Answer:** No. `go test ./...` currently has 93 passing and 2 failing packages because auth/workspaces test DB setup uses SSL-incompatible settings. Frontend typecheck/tests/build pass, with MSW warnings and an 848 kB bundle warning. I would fix backend reproducibility before submission.

### Missing Tests To Acknowledge

- Lock-safe refresh rotation that fails before the fix.
- Disabled-user refresh rejection.
- Strict impossible-date, invalid-order, invalid-pagination, and invalid-source validation.
- Centralized 403/422/429/500 frontend behavior.
- Dashboard/forecast/scenario page integration.
- Playwright critical flow.
- Mobile keyboard/focus/accessibility interactions.

## Performance & Scaling Q&A

### Basic: What would change at one million users?

**Answer:** Keep the API horizontally scalable, add read replicas for analytics, measure query plans, materialize/rebuild balance projections where justified, partition high-volume transactions by workspace/date, scale workers independently, isolate AI jobs and cost limits, and improve observability. Do not add these without workload evidence.

### Challenge: What if there are 10 million transactions?

**Answer:** Preserve bounded pagination and DB aggregation, inspect `EXPLAIN ANALYZE`, add/adjust date/workspace indexes, consider partitioning and precomputed analytics projections. A mutable balance shortcut would not be the first answer because it risks authority drift.

### Current Performance Caveat

The frontend production bundle is approximately 848 kB minified. Route-level code splitting is a reasonable next step, not a core correctness issue.

## Infrastructure Q&A

### Basic: Why Redis/worker?

**Answer:** Async notification/background processing and distributed coordination/rate limiting, while financial writes remain synchronous and authoritative. Worker jobs call shared services rather than duplicate domain logic.

### Challenge: What if Redis goes down?

**Answer:** Financial correctness must not change. Async features can degrade according to policy; security-sensitive rate limiting needs an explicit fail-open/closed decision. The application should report readiness clearly.

### Challenge: What if PostgreSQL goes down?

**Answer:** Financial reads/writes fail safely; no success response should imply a committed financial change. Readiness should fail for the critical DB dependency. Health should still distinguish process liveness from dependency readiness.

### Challenge: What does Docker Compose run?

**Answer:** Current Compose runs PostgreSQL, Redis, MinIO, and initialization infrastructure. It does not currently run API/frontend/worker services, so README language must not imply a complete containerized app.

## Trade-Off Q&A

| Decision | Why | Alternative | Why Not | Current Cost | Revisit When |
| --- | --- | --- | --- | --- | --- |
| Modular monolith | Transactions and reviewability | Microservices | Distributed transactions/ops too early | Coarser deployment scaling | Measured module isolation/scale need |
| Derived balance | Rebuildable authority | Mutable `current_balance` | Drift risk | More aggregation work | High-volume balance reads |
| BIGINT minor units | Exact money | Float/decimal everywhere | Float errors; more complexity | Currency scale handling | Multi-currency/FX P1 |
| Workspace tenancy | Collaboration and shared scope | User-owned rows only | Blocks household/team model | Membership checks | Larger tenant isolation need |
| Cookie auth | HttpOnly credentials | localStorage bearer tokens | XSS exposure | CSRF required | Different client trust model |
| Hashed refresh sessions | Revocation/rotation | Stateless refresh JWT | Harder revocation/replay control | DB session operations | Large auth scale with measured need |
| Redis worker | Async boundary | Synchronous everything | Slow/non-critical work blocks requests | Queue operations | Higher throughput/job isolation |
| Deterministic forecast | Reproducible finance | LLM prediction | Not authoritative or explainable enough | More domain formulas | Better statistical model/data |
| Non-destructive scenarios | User safety | Clone/mutate real records | Corrupts financial history | Snapshot/stale logic | Advanced planning model |
| Bounded AI tools | Least privilege | Arbitrary agent/SQL | Authorization and injection risk | Tool maintenance | Concrete new bounded use case |
| Application tenancy | Simpler P0 delivery | PostgreSQL RLS | More DB policy complexity | Every query must scope | Stronger DB isolation requirement |

## Failure Scenario Answers

### PostgreSQL goes down

Return safe dependency errors; do not claim financial writes succeeded. Readiness reports DB failure. No raw SQL/stack trace leaks.

### Redis goes down

Finance services remain authoritative and should not silently change correctness. Queue/rate-limit behavior follows explicit degradation policy.

### AI provider goes down

AI fails gracefully; analytics, forecast, scenarios, transactions, and budgets continue.

### Refresh token stolen

Opaque token is HttpOnly and DB stores hash. Rotation/revocation should invalidate the old token. Current concurrency sequencing is a known defect; fix lock-first hash validation and consider session-family revocation for stronger theft response.

### Old refresh token reused

Reject it; do not issue a new session. Log/audit safely without token contents. Current implementation needs stronger race-safe enforcement.

### Two users update a budget

Version check returns 409 for stale update. UI should offer reload rather than overwrite.

### Two workers confirm one occurrence

Unique occurrence protection plus atomic status transition allows one winner. Verify row locking and database constraint behavior in the exact migration.

### Two concurrent transfers

Atomic DB transactions protect partial state; use consistent account lock order if row locks are involved to reduce deadlocks. Verify with integration concurrency tests.

### Frontend retries after network failure

Safe reads may retry according to client policy. Do not automatically retry non-idempotent financial writes without an idempotency key/explicit model.

### Scenario viewed after finance changes

Mark stale via source revision and require recalculation.

### Invalid AI JSON

Reject schema-invalid output; return safe degraded error; never persist or execute it.

### Unknown AI tool

Allowlist rejects it. No arbitrary SQL, shell, filesystem, or HTTP.

### Guessed transaction UUID

Workspace-scoped lookup returns not found/denies access without revealing another workspace’s record.

## Challenge Questions

1. Why not PostgreSQL RLS? **P0 uses application-layer workspace scoping for simpler migrations and explicit service authorization; RLS is a future defense-in-depth option.**
2. What if a handler forgets authorization? **The service/repository boundary must still scope resources; tests target cross-workspace access. The review should identify any endpoint that violates this.**
3. Why can’t AI choose `workspace_id`? **Authorization context comes from authenticated backend state, not model output.**
4. What happens when `order=sideways`? **Current code silently treats non-`asc` as descending; this is a validation defect to fix, not defend.**
5. What happens on `2026-02-31`? **Current validator may accept it; strict `time.Parse` should replace the hand-rolled check.**
6. Why does the test DB fail? **Auth/workspaces test DSNs disagree with local PostgreSQL SSL mode; configuration must be centralized.**
7. Why no OpenAPI file? **Markdown API contract satisfies the assignment minimum, but internal docs claim OpenAPI; resolve the contradiction before submission.**
8. Why no mobile drawer? **Current responsive shell hides the sidebar below `md`; mobile navigation is incomplete.**
9. Why no frontend sorting? **Backend allowlist exists, but UI/API controls are missing.**
10. Can AI write a transaction? **No silent writes. Any future proposal needs normal validation plus explicit user confirmation.**
11. Why not use floats for percentages? **Percentage metadata is not authoritative money, but money fields remain integer-based.**
12. Why not AutoMigrate? **It cannot provide reviewed, reversible production schema history.**
13. What is the hardest correctness problem? **Separating financial authority from projections/AI while preserving immutable history and atomic workflows.**
14. What is the hardest security problem? **Cookie auth requires coordinated CSRF, refresh rotation, session revocation, and frontend concurrency handling.**
15. What if historical data is voided? **Keep the record for audit; exclude it from active analytics and recompute derived results.**
16. How do you stop transfer from inflating cashflow? **Transfers have separate records/effects and are excluded from income/expense aggregation.**
17. What if an account is archived? **Keep historical records, reject new ordinary activity.**
18. What would you instrument? **Request ID, duration/status/error code, DB query latency, queue age/failures, forecast duration, AI latency/cost, and authorization denials.**
19. What would you cache? **Measured read-heavy analytics or projections, never as an untracked authority for financial state.**
20. How would you control AI cost? **Per-user/workspace rate limits, bounded context, token budgets, provider timeouts, mock tests, and async work where appropriate.**

## Interview Risk Areas

| Potential Question | Risk | Safe Accurate Explanation | What Not To Claim |
| --- | --- | --- | --- |
| Is refresh rotation race-safe? | Actual backend gap | Frontend single-flight is strong; backend lock/hash sequencing needs fixing. | “Row lock makes it fully safe.” |
| Are all tests green? | Reviewer can reproduce failure | 93 pass, 2 fail from SSL DSN mismatch; fix before submission. | “CI is green.” |
| Are secrets fail-fast? | Weak secret acceptance | Current code logs weak/missing values; startup rejection is required. | “Invalid secrets cannot start.” |
| Does OpenAPI exist? | Missing artifact/internal contradiction | Markdown contract exists; `openapi.yaml` is absent. | “OpenAPI is complete.” |
| Does Docker run Savio? | Compose overclaim | Compose runs infra; not API/frontend/worker. | “`docker compose up` runs the full app.” |
| Is MinIO implemented? | Infrastructure vs feature confusion | MinIO is provisioned; no demonstrated upload workflow. | “File storage feature is complete.” |
| Is frontend sorting complete? | Backend/UI mismatch | Backend validates sort; frontend controls are missing. | “All listing requirements are complete.” |
| Is responsive UX complete? | Mobile quality gap | Desktop/tablet foundation exists; mobile nav/table/focus need work. | “All breakpoints are production-ready.” |
| Does frontend handle 403/422/429/500 centrally? | Assignment gap | Errors normalize centrally, status-specific UX is incomplete. | “Interceptor fully handles every status.” |
| Does Savio use RLS/tracing/double-entry? | Common overclaim trap | Application scoping, request IDs, and simpler ledger model are actual choices. | Claiming RLS, OpenTelemetry, or double-entry. |

## What I Would Improve

1. Fix refresh rotation with lock-first hash comparison, disabled-user check, and deterministic concurrency regression test.
2. Centralize test DSN/configuration, align CI Go version, and make `go test ./...` green from a fresh environment.
3. Add strict input validation and centralized Axios handling for 403/422/429/500.
4. Finish responsive mobile navigation/table/accessibility and add critical Playwright flow.
5. Resolve OpenAPI/README/Compose/implementation-progress contradictions.

These were not P0 feature additions. They improve submission correctness, trust, and delivery readiness.

## What Was Hardest

The hardest part was maintaining authority boundaries across financial records, derived balances, forecast/scenario projections, and AI interpretation. The second hardest was cookie authentication: HttpOnly credentials simplify token exposure but require CSRF, refresh rotation, server sessions, and frontend concurrent-request coordination.

## What I Am Most Proud Of

The strongest engineering decision is the explicit rule: “Finance Engine calculates. AI interprets. User decides.” It appears in the data model, service boundaries, forecast/scenario design, AI tools, and UI labeling. That makes the product safer and easier to defend than an AI system that invents financial numbers.

## Interview Cheat Sheet

### Product

Personal cashflow intelligence: understand today, see what comes next, test decisions before making them.

### Architecture

React → Go/Gin API → services/domain rules → PostgreSQL; Redis/worker for async work; AI only through bounded tools.

### Security

HttpOnly cookie auth + signed CSRF + server refresh sessions + backend workspace/RBAC checks; no token storage in browser storage.

### Finance Engine

Integer minor units, immutable posted history, derived balances, explicit lifecycle, atomic transfers.

### Forecast

Deterministic projection from current state, known/scheduled events, estimates, and assumptions.

### Scenario

Baseline forecast plus hypothetical modifications, never real-ledger mutation, stale when source revision changes.

### AI

Bounded tools return deterministic facts; LLM interprets; invalid output is rejected; provider failure is degradable.

### Five strongest decisions

1. Modular monolith.
2. Derived account balances.
3. Integer minor-unit money.
4. Non-destructive scenarios.
5. AI downstream of deterministic finance services.

### Five strongest tests

1. Cross-workspace IDOR.
2. Last-owner invariant.
3. Transaction lifecycle/voiding.
4. Recurring idempotency.
5. Frontend single-flight refresh.

### Five key trade-offs

1. Modular monolith over microservices.
2. Application tenancy over RLS.
3. Simple ledger over double-entry accounting.
4. Deterministic forecast over LLM prediction.
5. P0 breadth over complete E2E/runtime packaging.

## Mock Interview Questions

### 30 High-Probability Questions

1. Why did you choose Savio’s domain?
2. Who is the target user?
3. Why is Savio not CRUD?
4. Why modular monolith?
5. Why Go, Gin, GORM, and PostgreSQL?
6. Why explicit migrations?
7. Why integer minor units?
8. How is balance derived?
9. Why immutable posted transactions?
10. How do corrections work?
11. How are transfers atomic?
12. Why are transfers excluded from analytics?
13. How does recurring confirmation work?
14. How do you prevent duplicate recurring posts?
15. Why workspace-scoped resources?
16. How do OWNER/MEMBER/VIEWER permissions work?
17. How do you prevent IDOR?
18. Why cookie authentication?
19. How does CSRF work?
20. How does refresh rotation work?
21. What happens with concurrent refreshes?
22. How do you handle session revocation?
23. How does optimistic locking work?
24. Why is forecast deterministic?
25. What assumptions drive forecast?
26. How does stale forecast/scenario detection work?
27. Why are scenarios non-destructive?
28. How does AI access data?
29. What happens when AI is unavailable?
30. What would you improve with another week?

### 20 Deep / Challenge Follow-Ups

1. Prove the refresh rotation lock prevents two successes.
2. What exact SQL/index protects recurring idempotency?
3. What if an old refresh token is replayed after a race?
4. What if the user is disabled while an access token remains valid?
5. Why does UUID not solve IDOR?
6. Why do you need CSRF with SameSite cookies?
7. What happens if `2026-02-31` is submitted?
8. What happens if `order=sideways` is submitted?
9. Why are invalid pagination values normalized?
10. What if a category belongs to another workspace?
11. What if a transfer source equals destination?
12. How do you ensure a voided transaction never affects analytics?
13. What happens if forecast inputs change during calculation?
14. How do you prove scenario calculation did not mutate ledger state?
15. How does AI prompt injection fail safely?
16. Can the model choose workspace or user IDs?
17. What happens when AI returns unknown structured action?
18. Why does the frontend not handle 403/422/429/500 centrally?
19. How would you run the project with Docker on a clean machine?
20. At what measured point would you split the monolith?

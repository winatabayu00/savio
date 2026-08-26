# Savio — Take-Home Evaluation

## Executive Summary

Savio is an ambitious, coherent personal cashflow intelligence product rather than a CRUD expense tracker. The repository shows strong domain modeling, workspace authorization, deterministic forecast/scenario boundaries, explicit migrations, and meaningful financial/security tests.

The submission is not fully ready for an unconditional pass. The backend suite currently fails in `auth` and `workspaces` because integration tests use an SSL-incompatible DSN. More seriously, refresh rotation is not protected by a locked old-token hash check, so concurrent refresh requests can both pass the pre-lock lookup. Configuration accepts weak or missing JWT/CSRF secrets. The frontend lacks required status-specific Axios UX, listing sort controls, complete mobile navigation, and broad page tests. Docker Compose does not run the application services, and `docs/api/openapi.yaml` is absent despite internal documentation references.

## Reviewer Verdict

**BORDERLINE PASS**

Pass recommendation is conditional on fixing the refresh-rotation race and making the test suite reproducibly green. The core engineering judgment is strong enough to continue the candidate to interview, but security and delivery-readiness defects prevent a clean pass.

## Overall Score

| Category | Score | Maximum | Assessment |
| --- | ---: | ---: | --- |
| Backend Architecture & Code Quality | 12 | 15 | Clear modular monolith; some context/clock leakage and security defects. |
| API & Business Logic | 12 | 15 | Substantial financial workflow; validation edge cases remain. |
| Authentication & Security | 10 | 15 | Strong cookie/CSRF/RBAC foundation; refresh race and secret validation are serious. |
| UI/UX | 10 | 15 | Intentional, coherent product UI; mobile/accessibility/status gaps. |
| Database & Migration | 9 | 10 | Strong constraints, indexes, up/down migrations; test DSN reproducibility issue. |
| Frontend Architecture | 8 | 10 | Good feature structure and query/auth foundations; interceptor/RBAC/sort gaps. |
| Testing | 3 | 5 | Strong backend domain tests and auth tests; suite fails, shallow frontend coverage, no E2E. |
| Documentation | 3 | 5 | Extensive docs and API contract; missing OpenAPI and contradictory paths/status. |
| Git & Engineering Practice | 4 | 5 | Focused commits, clean tracked secrets; untracked test and CI Go-version mismatch. |
| **Base Score** | **71** | **100** |  |

### Bonus Assessment

Bonus engineering is substantial but uneven:

- Positive: polling worker, MinIO local infrastructure, AI provider abstraction, audit logging, request IDs, rate limiting, health/readiness, race-detector CI, migration tests, optimistic-locking/domain conflict handling.
- Partial: MinIO has no demonstrated implemented upload flow; Compose contains infrastructure but not API/frontend/worker; no Playwright critical-flow E2E; no OpenAPI artifact; CI declares Go 1.24 while `backend/go.mod` requires Go 1.27.
- Recommended reviewer treatment: meaningful differentiator, not score compensation for failed core security/tests.

## Requirement Matrix

Status uses the assignment vocabulary: `PASS`, `PASS WITH NOTES`, `PARTIAL`, `FAIL`, `NOT APPLICABLE`.

| Requirement | Expected Behavior | Implementation Evidence | Relevant Files / Endpoint / UI | Relevant Tests | Status | Reviewer Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Fullstack application | Backend, frontend, DB, auth | Go API, React app, PostgreSQL migrations, auth modules | `backend/`, `frontend/`, `backend/internal/migrations/` | Backend/frontend suites | PASS WITH NOTES | Full stack exists; Compose does not run app services. |
| Clear problem/user/business flow | Defined target and non-CRUD workflow | Track → Understand → Forecast → Simulate → Explain → Decide | `README.md`, product docs, dashboard/forecast/scenarios | Domain service tests | PASS | Product concept is immediately understandable. |
| Go backend | Go implementation | `go.mod`, `cmd/api` | `backend/go.mod` | `go test ./...` | PASS |  |
| Gin | Gin HTTP framework | Router/module wiring | `backend/internal/server/modules.go` | Handler tests | PASS |  |
| GORM | ORM | Repositories and DB setup | `backend/internal/*/repository.go` | Integration tests | PASS |  |
| PostgreSQL | PostgreSQL persistence | SQL migrations, pq driver | `backend/internal/migrations/` | Migration test | PASS WITH NOTES | Local integration test DSNs are inconsistent. |
| Maintainable architecture | Clear, navigable separation | Handler → service → repository modules | `backend/internal/*` | Module tests | PASS | Domain boundaries are reviewer-friendly. |
| React/TypeScript | React or Vue frontend | React TS feature pages | `frontend/package.json`, `frontend/src/` | Typecheck/build | PASS |  |
| Tailwind | Tailwind styling | Tailwind config/global styles/classes | `frontend/tailwind.config.ts`, `frontend/src/` | Build | PASS |  |
| Axios | Central HTTP client | One Axios instance with credentials/interceptors | `frontend/src/shared/api/client.ts` | `session-refresh.test.ts` | PASS WITH NOTES | Core refresh works client-side; status-specific UX absent. |
| Cookie authentication | Credentials in cookies | HttpOnly access/refresh cookies | `backend/internal/auth/cookies.go` | Auth tests | PASS | No browser token storage found. |
| No local/session storage auth | Never store auth tokens there | Auth uses cookies; grep found no token storage | `frontend/src/` | Auth tests | PASS | JSDOM emits unrelated localStorage availability warnings. |
| Login/logout/session expiry/current user | Complete auth lifecycle | Auth handlers, middleware, `/auth/me`, frontend provider | `backend/internal/auth/`, `frontend/src/features/auth/` | Auth flow tests | PASS WITH NOTES | Refresh failure path exists; backend refresh security defect remains. |
| Cookie attributes | HttpOnly, Secure, SameSite, Path, lifetime | Cookie helper config | `backend/internal/auth/cookies.go` | Auth tests | PASS WITH NOTES | Environment-dependent Secure behavior must be defended. |
| CSRF | Mutations require CSRF token | Signed double-submit token and global mutation middleware | `backend/internal/auth/csrf/`, `/api/v1/auth/csrf` | CSRF tests | PASS | Good explanation boundary. |
| RBAC/backend authorization | Multiple roles, enforced server-side | OWNER/MEMBER/VIEWER; write checks and membership scope | `backend/internal/auth/middleware.go`, workspace service | RBAC/last-owner/IDOR tests | PASS | Frontend role-aware UX is missing, but backend enforcement is correct. |
| Relationships/constraints/indexes | Integrity protected in DB | FKs, checks, unique constraints, indexes | `backend/internal/migrations/*.sql` | Migration lifecycle test | PASS | Defendable schema for P0. |
| Explicit migrations | Up/down, fresh reproducibility | Embedded golang-migrate source and down files | `backend/internal/migrations/` | `migrations_test.go` | PASS WITH NOTES | Test setup is not environment-portable. |
| Business workflow | Backend state transitions/rules | Draft → Posted → Voided, transfers, recurring, budgets, scenarios | `backend/internal/transactions/service.go`, modules | Finance module tests | PASS | Strongest part of submission. |
| Input validation | Body/path/query/auth fields validated server-side | UUID/money/enums/pagination/sort validation | Handlers and `httpx` | Validation tests | PARTIAL | Impossible dates accepted; invalid order/pagination normalized; `Source` not allowlisted. |
| Backend source of truth | Frontend not authoritative for finance | Analytics/forecast/scenario computed by backend | Backend services and API pages | Finance tests | PASS | Aligns with product thesis. |
| REST `/api/v1` | Versioned routes and status/envelopes | Versioned router, `httpx` envelopes | `backend/internal/server/modules.go` | Handler tests | PASS WITH NOTES | API contract exists; OpenAPI file absent. |
| Axios 401 | Single refresh, one retry, failure logout | `refreshPromise`, `_retried`, unauthorized event | `frontend/src/shared/api/client.ts` | Refresh tests | PASS WITH NOTES | Frontend behavior strong; backend rotation race invalidates full security claim. |
| Axios 403/422/429/500 | Status-specific feedback | Generic `toApiError` only | `frontend/src/shared/api/client.ts` | No status interceptor tests | FAIL | Pages may handle some errors locally, but required interceptor behavior is not implemented. |
| Search/filter/sort/pagination | Listing supports all four | Transactions API/backend supports search/filter/pagination and sort allowlist | Transactions page/API/handler | Transaction listing tests | PARTIAL | Backend sort exists; frontend has no sort controls/params. |
| UI/UX states | Loading, disabled, empty, error, form feedback, destructive confirmation, responsive | Feature pages, modal, empty-state, forms | `frontend/src/features/`, `shared/components/ui/` | Limited auth/voice tests | PARTIAL | States exist, but mostly plain loading text, no reusable toast, mobile table/sidebar/accessibility gaps. |
| Reusable components | Avoid duplicate UI primitives | Button, text field, modal, empty state | `frontend/src/shared/components/ui/` | No component suite | PARTIAL | Useful foundation; requested breadth not met. |
| Centralized error handling | Safe backend errors and frontend normalization | `errs`, recovery middleware, `ApiError` | `backend/internal/platform/errs/`, `client.ts` | Error tests | PASS WITH NOTES | Frontend status UX incomplete. |
| API documentation | Endpoint/auth/request/query/response/errors/examples | Detailed Markdown API contract | `docs/api/api-contract.md` | N/A | PASS WITH NOTES | Assignment minimum permits Markdown; internal docs still reference missing `openapi.yaml`. |
| README/.env/Docker/migrations/source | Repository deliverables and setup | README, `.env.example`, Compose, migrations | Root files | Compose config | PASS WITH NOTES | Setup understandable; Compose is infra-only and docs contain path mismatches. |
| Backend testing | Unit, preferably integration | Extensive service tests | `backend/internal/**/*_test.go` | `go test ./...` | PARTIAL | 93 pass, 2 fail due SSL DSN mismatch. |
| Frontend testing | Component/integration; E2E bonus | 4 files, 21 tests | `frontend/tests/` | `npm run test -- --run` | PARTIAL | Auth/refresh/voice/errors covered; core pages and E2E missing. |
| File storage bonus | Actual secure object-storage flow | MinIO in Compose only | `docker-compose.yml` | None found | PARTIAL | Infrastructure is not a feature. |
| Worker bonus | Async processing with Redis/RabbitMQ | Direct PostgreSQL polling worker | `backend/cmd/worker`, `internal/worker` | Worker tests | PASS WITH NOTES | Scheduled jobs work; Redis queue processing is not implemented. |
| Docker bonus | `docker compose up` runs needed services | PostgreSQL/Redis/MinIO services | `docker-compose.yml` | `docker compose config --quiet` | PARTIAL | API/frontend/worker not included. |
| Creativity | Value-add feature quality | Analytics, forecast, scenario, AI, audit | Feature modules/pages | Module tests | PASS | Feature selection is coherent, not quantity-driven. |
| Performance/security/observability bonus | Indexes, rate limits, locking, logs, request ID, health | Present across platform/modules | `platform/`, CI, middleware | Related tests | PASS WITH NOTES | Refresh race, secret validation, no tracing/metrics. |

## Scoring Breakdown

### Backend Architecture — 12/15

**Evidence:** Modular domain packages with thin handlers, service-owned workflows, repositories for DB access, platform packages for shared concerns.

**Strengths:** Clear module boundaries; finance, auth, AI, worker dependency direction is defensible; no premature microservices.

**Deductions:** Request cancellation is weakened by `context.Background()` in some paths. Injectable clock discipline is incomplete. Refresh security defect crosses architecture and correctness boundaries.

### API & Business Logic — 12/15

**Evidence:** Transaction lifecycle, transfer invariants, recurring occurrence handling, budgets, deterministic forecast, scenarios, workspace scoping.

**Strengths:** Clearly non-trivial domain; posted transactions are immutable; corrections use void/replacement semantics; scenarios are non-destructive.

**Deductions:** Strict validation gaps: impossible dates, silently normalized pagination/order, non-allowlisted transaction source. API documentation is not fully aligned with intended OpenAPI source.

### Authentication & Security — 10/15

**Evidence:** HttpOnly cookies, signed CSRF, Argon2, rate limits, server-side sessions, roles, IDOR tests.

**Strengths:** Security model is unusually explicit for a take-home. Backend authorization and CSRF are real controls, not UI claims.

**Deductions:** Refresh rotation currently looks up the token before locking and does not re-check the old hash inside the lock. Weak/missing JWT and CSRF secrets are logged rather than rejected. Refresh does not clearly revalidate disabled-user status. These are interview-grade concerns, not cosmetic deductions.

### UI/UX — 10/15

**Evidence:** Product-specific dashboard, transactions, forecast, scenarios, insights, Copilot; forms and confirmation modal; calm actual/projected terminology.

**Strengths:** UI supports the product story and does not make AI the authority. Empty/error/form states exist.

**Deductions:** Sidebar disappears without a mobile replacement; transaction table remains a table on mobile; no focus trap/restoration in modal; no visible sort UI; loading and success feedback are not consistently componentized.

### Database & Migration — 9/10

**Evidence:** Explicit embedded up/down migrations; FK/check/unique/index coverage; integer minor units; workspace-scoped relationships.

**Strengths:** Schema is defendable and financial history remains reconstructable.

**Deduction:** Fresh migration/integration verification is not portable because test DSNs are hardcoded/inconsistent with local SSL settings.

### Frontend Architecture — 8/10

**Evidence:** Feature folders, central Axios, TanStack Query, React Hook Form/Zod, auth bootstrap unknown state.

**Strengths:** Good state ownership and API boundary; no large unnecessary global store.

**Deductions:** Axios does not implement required status-specific behavior; frontend has no role-aware controls; core listing sorting is absent; route-level bundle splitting is absent.

### Testing — 3/5

**Evidence:** Strong finance/auth/RBAC/CSRF/refresh/recurring tests; 21 frontend tests.

**Strengths:** Tests target invariants, not only happy paths. IDOR, last-owner, refresh replay/concurrency intent, and financial lifecycle tests would impress a reviewer.

**Deductions:** `go test ./...` fails. Frontend coverage omits core business pages, error statuses, RBAC UI, responsive interactions, and critical E2E flow. Existing refresh race test does not protect the implementation from the race it intends to test.

### Documentation — 3/5

**Evidence:** README, product, architecture, security, database, API, testing, implementation docs.

**Strengths:** Product and technical decisions are unusually well explained.

**Deductions:** Missing `docs/api/openapi.yaml`. Tailwind-first frontend and production Docker deployment remain unimplemented.

### Git Quality — 4/5

**Evidence:** Recent focused conventional commits such as `fix(auth)`, `fix(money)`, `feat(ai)`, `fix(scenarios)`.

**Strengths:** History shows iterative hardening, not one giant final commit. `.env` is not tracked; `git diff --check` is clean.

**Deduction:** `frontend/tests/unit/voice.test.tsx` is untracked. CI specifies Go 1.24 while the module requires Go 1.27.

## Product Evaluation

Savio feels like personal cashflow intelligence, not “expense tracker + AI chat.” The core flow is connected:

1. Track accounts, transactions, transfers, recurring activity, budgets, and goals.
2. Understand through analytics and category/period breakdowns.
3. Forecast using deterministic assumptions and future events.
4. Simulate decisions without mutating actual records.
5. Explain results through bounded AI interpretation.
6. Decide with actual, projected, scenario, and AI-generated states separated.

The differentiator is authority separation: PostgreSQL and deterministic services calculate; AI interprets; the user decides. The product risk is breadth. The implementation has many modules for a take-home, so the candidate must explain that the modules form one demo story rather than independent feature accumulation.

## Backend Architecture

**Good decisions:** modular monolith; handlers bind and map; services own authorization and workflows; repositories own queries; shared platform packages handle money, errors, auth context, rate limiting, and middleware; worker calls shared services.

**Underengineering:** refresh concurrency protection is incomplete; some request context and clock paths are not disciplined; configuration validation is permissive.

**Overengineering risk:** documentation and feature breadth exceed the current runtime packaging. MinIO/OpenAPI/worker claims should be described precisely rather than presented as complete production integrations.

## Business Logic

Strongest implementation-backed examples:

1. `DRAFT → POSTED → VOIDED`, with posted financial fields immutable.
2. Corrections through void plus replacement rather than destructive editing.
3. Account balance reconstructed from opening balance and posted ledger effects.
4. Atomic transfers between distinct same-workspace active accounts, excluded from income/expense analytics.
5. Recurring occurrences confirmed at most once through domain/database protections.
6. OWNER/MEMBER/VIEWER permissions, including last-owner protection.
7. Budget actuals derived from posted expense transactions matching category and period.
8. Deterministic forecast with known/scheduled/estimated/assumed events and freshness.
9. Scenario overlays that preserve actual records and become stale after source finance changes.
10. Workspace-scoped resource lookup preventing UUID-only IDOR.

## Authentication & Security

The design is interview-worthy: access and refresh cookies, hashed refresh sessions, signed double-submit CSRF, backend role/scope checks, Argon2 passwords, rate limiting, security headers, and safe errors.

The candidate must not overclaim refresh rotation. Actual code performs a pre-lock lookup and later locks by session ID without verifying that the token hash is still the expected old hash. The correct defense is to lock first, re-read, compare hash inside the transaction, rotate once, and reject stale reuse. Weak secrets must fail startup, not merely produce logs.

## Database & Migrations

The schema is defendable for P0: workspace tenancy, explicit memberships, financial entities, constraints, indexes, and integer minor units. It intentionally does not implement full double-entry accounting. The key delivery weakness is reproducibility: migration tests fail in the current environment because test connection settings disagree on SSL.

## Frontend Architecture

Feature organization and server-state management are strong. `AuthProvider` preserves an unknown bootstrap state. The Axios client correctly uses `withCredentials`, CSRF on mutations, single-flight refresh, one retry, and an unauthorized event.

The assignment's interceptor requirement is not fully met. `403`, `422`, `429`, and `500` are normalized into `ApiError`, but there is no centralized status-specific behavior. Sort is supported by backend allowlists but not exposed in the frontend listing UI. Role-aware controls are absent, which is acceptable for backend security but weaker UX.

## UI / UX

The UI is intentionally product-specific and coherent. It distinguishes actual/projected/scenario/AI information and includes major product surfaces. Reviewer concerns: mobile sidebar replacement, mobile transaction cards/lists, modal focus management, keyboard interaction for clickable rows, reusable toast/loading components, and more complete error/loading feedback.

## Testing

Most impressive tests target financial and security invariants: transaction lifecycle/voiding, transfer behavior, recurring idempotency, last-owner protection, RBAC, IDOR, CSRF, refresh replay/concurrency intent, and frontend single-flight refresh.

Missing or weak coverage: a truly race-safe refresh implementation test, disabled-user refresh, strict date/order/pagination/source validation, all status-specific Axios paths, dashboard/forecast/scenario page integration, responsive critical interactions, and a Playwright login → create data → forecast/scenario flow.

## Documentation

Documentation quality is a strength in volume and intent. It gives the reviewer a product, architecture, security, and testing narrative. It loses trust through contradictions: missing OpenAPI artifact, incorrect migration path, stale implementation-progress status, and README claims that exceed Compose/runtime packaging.

## Git Quality

History is focused and shows incremental security/financial fixes. No tracked secret was found. Clean the untracked voice test before submission and align CI Go version with `backend/go.mod`.

## Bonus Engineering

| Bonus | Assessment |
| --- | --- |
| Analytics | PASS; meaningful backend aggregation and UI. |
| Forecast | PASS; deterministic and explainable. |
| Scenario simulation | PASS; non-destructive and stale-aware. |
| AI/provider abstraction | PASS WITH NOTES; bounded architecture and mock provider; validate external claims carefully. |
| Audit | PASS WITH NOTES; present in backend, depth should be demonstrated. |
| Notifications | PASS WITH NOTES; schema/worker paths exist. |
| Queue/Redis | PASS WITH NOTES; worker exists; Compose runtime path incomplete. |
| MinIO | PARTIAL; infrastructure exists, implemented storage workflow not demonstrated. |
| Docker | PARTIAL; infrastructure Compose, not full application Compose. |
| Rate limiting/request ID/health/logging | PASS WITH NOTES; good baseline, no full metrics/tracing. |
| Optimistic locking | PASS WITH NOTES; domain support exists; frontend conflict UX breadth limited. |
| CI | PARTIAL; useful pipeline, Go-version mismatch can break it. |
| E2E | FAIL as bonus; no Playwright critical flow found. |

## Strongest Parts of the Submission

1. Clear authority boundary: Finance Engine calculates; AI interprets.
2. Non-trivial financial workflow with immutable posted history and void/replacement correction.
3. Workspace-scoped authorization with roles, IDOR tests, and last-owner invariant.
4. Deterministic forecast and non-destructive scenario model.
5. Focused iterative Git history and extensive engineering documentation.

## Weakest Parts of the Submission

1. Refresh rotation is not concurrency-safe in actual implementation.
2. Backend test suite fails in two packages.
3. Secret validation, strict date/order/pagination/source validation are incomplete.
4. Frontend does not fulfill status-specific Axios UX or listing sorting.
5. Runtime/docs packaging overstates OpenAPI, Docker, MinIO, and implementation completion.

## Blockers

### BLOCKER — Technical defect

Refresh token rotation can accept concurrent stale refresh requests because old-token validation occurs before the lock and is not repeated inside the locked transaction. Fix before submission.

### BLOCKER — Delivery defect

`go test ./...` is not green: `auth` and `workspaces` panic with `pq: SSL is not enabled on the server`. Fix DSN configuration and replace environment panics with clean test setup behavior.

## High-Priority Reviewer Concerns

- **HIGH / Security:** Weak or missing `JWT_SECRET` and `CSRF_SECRET` are logged, not rejected.
- **HIGH / Correctness:** Refresh should revalidate disabled-user/session state.
- **HIGH / Validation:** Strict date parsing and explicit invalid query rejection are needed.
- **HIGH / Documentation:** Align CI Go version, migration path, OpenAPI claims, and progress status.
- **MEDIUM / Frontend:** Add status-specific 403/422/429/500 behavior, sort controls, role-aware UX, mobile navigation/table treatment.
- **MEDIUM / Testing:** Add critical E2E and page-level integration tests.
- **LOW / Performance:** Split the 848 kB frontend bundle when product scale warrants it.

## Likely Reviewer Comments

- “This is substantially more than CRUD; the financial invariants are thoughtful.”
- “Show me the exact transaction boundary and prove refresh rotation is safe under concurrency.”
- “Why does the repository claim OpenAPI when the file is absent?”
- “How do I run the full application with Docker Compose?”
- “The backend supports sort. Why can’t the user sort the listing?”
- “The test suite failing on a fresh reviewer machine is a submission-readiness issue.”

## Why This Project Is More Than CRUD

CRUD stores records. Savio defines financial authority and transitions: posted history is immutable, voiding preserves auditability, transfers preserve portfolio value, recurring occurrences become actual at most once, budgets and analytics derive from authoritative records, forecasts use explicit deterministic assumptions, and scenarios overlay a baseline without mutation. Authorization and freshness are domain rules, not UI decoration.

## Submission Recommendation

**BORDERLINE PASS.** Continue to technical interview because the product and architecture demonstrate strong engineering judgment. Require the candidate to acknowledge and explain the refresh race, failed integration tests, and incomplete runtime/documentation claims. A clean pass follows quickly after fixing those core submission blockers; absent those fixes, downgrade to borderline fail.

# Savio — Database Design

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Business Requirements](../product/business-requirements.md) — business rules shaped into entities/constraints.
- [System Architecture](../architecture/system-architecture.md) — application context for the data model.
- [API Contract](../api/api-contract.md) — endpoints that read/write these tables.
- [Implementation Plan](../engineering/implementation-plan.md) — schema milestones and migrations.

## 1. Document Purpose

This document defines the initial relational database design for Savio.

The purpose of this document is to translate the product foundation, business requirements, and user flows into a concrete PostgreSQL schema that supports:

- secure user identity,
- financial accounts,
- transactions,
- transfers,
- recurring transactions,
- budgets,
- financial goals,
- cashflow analytics,
- forecast snapshots,
- financial scenarios,
- AI insights,
- notifications,
- auditability,
- and session management.

This document is intended to serve as the primary database design source of truth before implementation.

The core Savio principle remains:

> **Finance Engine calculates. AI interprets. User decides.**

The database stores authoritative financial state.

AI-generated data must never silently replace authoritative financial records.

---

# 2. Database Technology

Primary database:

```text
PostgreSQL
```

ORM:

```text
GORM
```

Migration strategy:

```text
Explicit SQL migrations
```

Recommended migration tool:

```text
golang-migrate
```

Production schema changes must not rely on GORM AutoMigrate as the source of truth.

---

# 3. Database Design Principles

The schema should follow these principles:

1. Use UUID primary keys for domain entities.
2. Use explicit foreign keys.
3. Use database constraints where appropriate.
4. Preserve financial history.
5. Avoid floating-point financial arithmetic.
6. Separate real financial data from projections and simulations.
7. Scope financial resources by `workspace_id` (see section 3.1). Users are
   keyed by `id`; financial entities belong to a workspace, and audit fields
   record `created_by_user_id` for actor identity.
8. Use archive/status semantics when deletion would destroy financial history.
9. Use transactional writes for balance-affecting operations.
10. Support optimistic concurrency for important mutable entities.
11. Keep AI-generated data separate from authoritative financial records.
12. Make indexes intentional and query-driven.
13. Preserve auditability for important financial actions.
14. Store timestamps in UTC.
15. Preserve user-local financial dates separately where necessary.

## 3.1 Workspace Scoping Correction

Some earlier sections below were authored before the workspace authorization
decision and reference `user_id` on financial tables. The implemented schema
follows the workspace model in `docs/engineering/implementation-plan.md`
(sections 9–13) and `AGENTS.md`:

```text
Financial resources are scoped by workspace_id:
accounts.workspace_id
transactions.workspace_id
transfers.workspace_id
categories.workspace_id  (NULL for system categories)
recurring_rules.workspace_id
budgets.workspace_id
goals.workspace_id
scenarios.workspace_id
forecast_snapshots.workspace_id
ai_insights.workspace_id
```

Identity tables remain user-scoped:

```text
users             (user.id)
auth_sessions     (user_id)
user_settings     (user_id)
```

Ownership of a financial row is proven by a row-scoped query such as:

```sql
WHERE workspace_id = $1 AND id = $2
```

plus membership verification in service code. A valid foreign key alone is
not ownership. Actor identity is preserved via `created_by_user_id`.

Within a single-user default workspace this preserves personal-finance
simplicity while still supporting backend-enforced multi-role authorization.

---

# 4. Identifier Strategy

Primary domain identifiers should use UUID.

Example:

```text
550e8400-e29b-41d4-a716-446655440000
```

Recommended PostgreSQL type:

```sql
UUID
```

IDs may be generated:

- in application code, or
- using PostgreSQL UUID generation.

The strategy should remain consistent.

---

# 5. Financial Amount Strategy

Financial calculations must not use floating-point types such as:

```text
FLOAT
REAL
DOUBLE PRECISION
```

for authoritative amounts.

For Savio MVP, recommended database type:

```sql
BIGINT
```

storing integer minor units.

Example:

```text
1200000000
```

represents 12,000,000.00 at a 2-decimal minor-unit scale.

This supports currencies that use decimal fractions.

Application code must use a decimal-safe representation over integer minor units.

API values travel as decimal-safe strings converted to/from minor units, so this decision applies consistently across the entire finance engine.

---

# 6. Currency Strategy

Initial MVP assumes:

```text
IDR
```

as the primary currency.

However, account and user records should still store currency codes for future extensibility.

Recommended type:

```text
CHAR(3)
```

Examples:

```text
IDR
USD
SGD
```

Initial business logic should reject unsupported cross-currency transfers.

---

# 7. Timestamp Strategy

Use:

```sql
TIMESTAMPTZ
```

for timestamps.

Examples:

```text
created_at
updated_at
deleted_at
generated_at
last_used_at
expires_at
```

Application timestamps should be stored in UTC.

Financial date fields that represent a user-selected calendar date may use:

```sql
DATE
```

Examples:

```text
transaction_date
start_date
target_date
occurrence_date
```

This prevents timezone conversion from changing the intended financial date.

---

# 8. High-Level Entity Relationship Model

The main relational structure is:

```text
users
│
├── auth_sessions
├── accounts
│   ├── transactions
│   ├── transfers
│   └── recurring_transactions
│
├── categories
│   ├── transactions
│   ├── recurring_transactions
│   └── budgets
│
├── budgets
├── financial_goals
├── recurring_transactions
├── forecast_snapshots
├── financial_scenarios
│   ├── scenario_modifications
│   └── scenario_snapshots
│
├── ai_insights
│   └── ai_insight_feedback
│
├── notifications
├── audit_logs
└── user_settings
│
├── telegram_settings
└── telegram_processed
```

---

# 9. Core Entity List

Initial schema includes:

```text
users
auth_sessions
user_settings

accounts
categories
transactions
transfers

recurring_transactions
recurring_occurrences

budgets
financial_goals

forecast_snapshots
forecast_events

financial_scenarios
scenario_modifications
scenario_snapshots

ai_insights
ai_insight_feedback
ai_requests

notifications
audit_logs
```

# 9a. Telegram Recap Storage

`telegram_settings` is keyed by `workspace_id` (PRIMARY KEY): every workspace
configures and owns its own bot row, so one workspace never blocks another:

- `enabled` — master switch; when off the worker skips polling entirely.
- `bot_token` — Telegram Bot API token (never returned unmasked by the API).
- `chat_id` — authorized chat; messages from any other chat are ignored.
- `workspace_id` → `workspaces(id)` — the workspace recap expenses are written to.
- `last_update_id` — best-effort Telegram long-poll offset.
- `webhook_url` / `webhook_secret` — optional push mode (registered webhook URL
  + the random path/header secret guarding the unauthenticated endpoint).

`telegram_processed` has a PRIMARY KEY `(workspace_id, update_id)` that guards
long-poll exactly-once per bot: an update is claimed
(INSERT ... ON CONFLICT DO NOTHING) before being handled, so a worker crash
between transaction creation and offset ack can never duplicate a transaction
(AGENTS #105). `update_id` is only unique per bot, which is why the composition
includes the workspace.

Recap writes go through the shared `transactions.Service` (AGENTS #102) as
POSTED expenses with `source = 'TELEGRAM'`.

Optional future tables:

```text
households
household_members

transaction_tags
tags

import_jobs
import_rows

attachments
receipts

merchant_mappings
```

---

# 10. Users Table

Table:

```text
users
```

Purpose:

Stores Savio user identity and high-level account state.

Suggested schema:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,

    name VARCHAR(120) NOT NULL,

    email VARCHAR(255) NOT NULL,

    password_hash VARCHAR(255) NOT NULL,

    timezone VARCHAR(100) NOT NULL DEFAULT 'Asia/Jakarta',

    default_currency CHAR(3) NOT NULL DEFAULT 'IDR',

    locale VARCHAR(20) NOT NULL DEFAULT 'id-ID',

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    email_verified_at TIMESTAMPTZ NULL,

    last_login_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Constraints:

```text
email unique
status enum-like constraint
default_currency length = 3
```

Recommended unique index:

```sql
CREATE UNIQUE INDEX users_email_lower_uq
ON users (LOWER(email));
```

Recommended status values:

```text
ACTIVE
DISABLED
```

---

# 11. Authentication Sessions

Table:

```text
auth_sessions
```

Purpose:

Tracks refresh sessions and device-level authentication state.

Suggested schema:

```sql
CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    refresh_token_hash VARCHAR(255) NOT NULL,

    user_agent TEXT NULL,

    ip_address INET NULL,

    device_name VARCHAR(150) NULL,

    expires_at TIMESTAMPTZ NOT NULL,

    last_used_at TIMESTAMPTZ NULL,

    revoked_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT auth_sessions_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Indexes:

```sql
CREATE INDEX auth_sessions_user_id_idx
ON auth_sessions(user_id);

CREATE INDEX auth_sessions_expires_at_idx
ON auth_sessions(expires_at);
```

The raw refresh token must never be stored.

Only a secure hash should be stored.

---

# 12. User Settings

Table:

```text
user_settings
```

Purpose:

Stores user-specific product preferences.

Suggested schema:

```sql
CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY,

    ai_insights_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    ai_copilot_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    budget_warning_threshold NUMERIC(5,2) NOT NULL DEFAULT 80.00,

    low_balance_threshold BIGINT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT user_settings_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Additional preferences may later be added through explicit migrations.

---

# 13. Accounts

Table:

```text
accounts
```

Purpose:

Represents locations where a user stores money.

Examples:

```text
Cash
Bank
E-Wallet
Savings
```

Suggested schema:

```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    name VARCHAR(120) NOT NULL,

    type VARCHAR(30) NOT NULL,

    currency CHAR(3) NOT NULL DEFAULT 'IDR',

    initial_balance BIGINT NOT NULL DEFAULT 0,

    current_balance BIGINT NOT NULL DEFAULT 0,

    institution_name VARCHAR(150) NULL,

    description TEXT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    version BIGINT NOT NULL DEFAULT 1,

    archived_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT accounts_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Supported type values:

```text
CASH
BANK
EWALLET
SAVINGS
OTHER
```

Supported status:

```text
ACTIVE
ARCHIVED
```

Indexes:

```sql
CREATE INDEX accounts_user_id_idx
ON accounts(user_id);

CREATE INDEX accounts_user_status_idx
ON accounts(user_id, status);
```

Optional user-scoped name uniqueness:

```sql
CREATE UNIQUE INDEX accounts_user_name_active_uq
ON accounts(user_id, LOWER(name))
WHERE status = 'ACTIVE';
```

---

# 14. Account Balance Model

Savio may persist `current_balance` for efficient reads.

However:

```text
current_balance
```

must be updated only through approved finance operations.

Conceptually:

```text
Current Balance
=
Initial Balance
+ Income
- Expense
+ Incoming Transfer
- Outgoing Transfer
+ Adjustment
```

Direct user editing of `current_balance` is prohibited.

Balance-affecting operations must use database transactions.

---

# 15. Categories

Table:

```text
categories
```

Purpose:

Classifies income and expense transactions.

Suggested schema:

```sql
CREATE TABLE categories (
    id UUID PRIMARY KEY,

    user_id UUID NULL,

    name VARCHAR(120) NOT NULL,

    type VARCHAR(20) NOT NULL,

    parent_id UUID NULL,

    is_system BOOLEAN NOT NULL DEFAULT FALSE,

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    icon VARCHAR(100) NULL,

    description TEXT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT categories_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT categories_parent_fk
        FOREIGN KEY (parent_id)
        REFERENCES categories(id)
        ON DELETE SET NULL
);
```

Supported category types:

```text
INCOME
EXPENSE
```

Supported status:

```text
ACTIVE
ARCHIVED
```

System categories:

```text
user_id = NULL
is_system = TRUE
```

Custom categories:

```text
user_id = user UUID
is_system = FALSE
```

---

# 16. Category Uniqueness

System category uniqueness:

```text
name + type
```

Custom category uniqueness:

```text
user_id + name + type
```

Because `NULL` behavior complicates one global constraint, separate partial indexes may be used.

Example:

```sql
CREATE UNIQUE INDEX categories_system_name_type_uq
ON categories(LOWER(name), type)
WHERE user_id IS NULL;

CREATE UNIQUE INDEX categories_user_name_type_uq
ON categories(user_id, LOWER(name), type)
WHERE user_id IS NOT NULL;
```

---

# 17. Transactions

Table:

```text
transactions
```

Purpose:

Stores authoritative income, expense, and adjustment records.

Transfers should preferably use the dedicated `transfers` entity rather than storing them as ordinary income/expense.

Suggested schema:

```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    account_id UUID NOT NULL,

    category_id UUID NULL,

    recurring_transaction_id UUID NULL,

    type VARCHAR(20) NOT NULL,

    amount BIGINT NOT NULL,

    transaction_date DATE NOT NULL,

    description VARCHAR(255) NULL,

    notes TEXT NULL,

    merchant VARCHAR(180) NULL,

    source VARCHAR(30) NOT NULL DEFAULT 'MANUAL',

    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',

    adjustment_reason TEXT NULL,

    version BIGINT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT transactions_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT transactions_account_fk
        FOREIGN KEY (account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT transactions_category_fk
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT
);
```

Supported types:

```text
INCOME
EXPENSE
ADJUSTMENT
```

Suggested status:

```text
DRAFT
POSTED
VOIDED
```

Suggested source:

```text
MANUAL
RECURRING
IMPORT
RECEIPT
SYSTEM
```

---

# 18. Transaction Constraints

Amount:

```sql
CHECK (amount > 0)
```

Example:

```sql
ALTER TABLE transactions
ADD CONSTRAINT transactions_amount_positive_ck
CHECK (amount > 0);
```

For `INCOME` and `EXPENSE`:

```text
category_id required
```

For `ADJUSTMENT`:

```text
adjustment_reason required
```

Some of these cross-field rules are easier to enforce in application service logic, but important invariants may also use database CHECK constraints where maintainable.

---

# 19. Transaction Ownership Integrity

Application queries must scope transactions by:

```text
transaction.user_id
```

The backend must additionally verify:

```text
account.user_id = transaction.user_id
```

and if category is user-owned:

```text
category.user_id = transaction.user_id
```

System categories are also valid.

A valid foreign key alone does not prove correct ownership.

---

# 20. Transaction Indexes

Common transaction queries:

```text
user + date
user + account
user + category
user + type
user + merchant
```

Recommended indexes:

```sql
CREATE INDEX transactions_user_date_idx
ON transactions(user_id, transaction_date DESC);

CREATE INDEX transactions_user_account_date_idx
ON transactions(user_id, account_id, transaction_date DESC);

CREATE INDEX transactions_user_category_date_idx
ON transactions(user_id, category_id, transaction_date DESC);

CREATE INDEX transactions_user_type_date_idx
ON transactions(user_id, type, transaction_date DESC);
```

Search requirements may later justify trigram or full-text indexes, but these should not be added prematurely.

---

# 21. Transfers

Table:

```text
transfers
```

Purpose:

Represents movement of money between two accounts owned by the same user.

Suggested schema:

```sql
CREATE TABLE transfers (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    source_account_id UUID NOT NULL,

    destination_account_id UUID NOT NULL,

    amount BIGINT NOT NULL,

    transfer_date DATE NOT NULL,

    description VARCHAR(255) NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'POSTED',

    version BIGINT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT transfers_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT transfers_source_account_fk
        FOREIGN KEY (source_account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT transfers_destination_account_fk
        FOREIGN KEY (destination_account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT transfers_different_accounts_ck
        CHECK (source_account_id <> destination_account_id),

    CONSTRAINT transfers_amount_positive_ck
        CHECK (amount > 0)
);
```

Status:

```text
POSTED
VOIDED
```

---

# 22. Transfer Indexes

```sql
CREATE INDEX transfers_user_date_idx
ON transfers(user_id, transfer_date DESC);

CREATE INDEX transfers_source_account_idx
ON transfers(source_account_id);

CREATE INDEX transfers_destination_account_idx
ON transfers(destination_account_id);
```

---

# 23. Transfer Transaction Boundary

Creating a transfer must atomically:

```text
1. verify ownership
2. verify account status
3. lock / safely update source account
4. lock / safely update destination account
5. decrement source balance
6. increment destination balance
7. create transfer record
8. write audit event
9. commit
```

If any step fails:

```text
ROLLBACK
```

Transfers must not affect:

```text
total income
total expense
savings rate
```

---

# 24. Recurring Transactions

Table:

```text
recurring_transactions
```

Purpose:

Stores recurring income and expense definitions.

Suggested schema:

```sql
CREATE TABLE recurring_transactions (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    account_id UUID NOT NULL,

    category_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,

    amount BIGINT NOT NULL,

    frequency VARCHAR(20) NOT NULL,

    interval_value INTEGER NOT NULL DEFAULT 1,

    start_date DATE NOT NULL,

    next_occurrence_date DATE NULL,

    end_date DATE NULL,

    day_of_month SMALLINT NULL,

    day_of_week SMALLINT NULL,

    auto_post BOOLEAN NOT NULL DEFAULT FALSE,

    description VARCHAR(255) NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    version BIGINT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT recurring_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT recurring_account_fk
        FOREIGN KEY (account_id)
        REFERENCES accounts(id)
        ON DELETE RESTRICT,

    CONSTRAINT recurring_category_fk
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT,

    CONSTRAINT recurring_amount_positive_ck
        CHECK (amount > 0),

    CONSTRAINT recurring_interval_positive_ck
        CHECK (interval_value > 0)
);
```

Supported types:

```text
INCOME
EXPENSE
```

Supported frequencies:

```text
DAILY
WEEKLY
MONTHLY
YEARLY
```

Supported statuses:

```text
ACTIVE
PAUSED
ENDED
```

---

# 25. Recurring Date Constraints

If `end_date` exists:

```text
end_date >= start_date
```

Example:

```sql
ALTER TABLE recurring_transactions
ADD CONSTRAINT recurring_valid_date_range_ck
CHECK (
    end_date IS NULL
    OR end_date >= start_date
);
```

For monthly recurrence:

```text
day_of_month
```

must be valid when used.

Application logic must define behavior for:

```text
29
30
31
```

on shorter months.

Recommended behavior:

```text
use last valid day of month
```

This must be documented and tested.

---

# 26. Recurring Occurrences

Table:

```text
recurring_occurrences
```

Purpose:

Tracks generated occurrences and prevents duplicate recurring posting.

Suggested schema:

```sql
CREATE TABLE recurring_occurrences (
    id UUID PRIMARY KEY,

    recurring_transaction_id UUID NOT NULL,

    user_id UUID NOT NULL,

    occurrence_date DATE NOT NULL,

    transaction_id UUID NULL,

    status VARCHAR(20) NOT NULL,

    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT recurring_occurrences_rule_fk
        FOREIGN KEY (recurring_transaction_id)
        REFERENCES recurring_transactions(id)
        ON DELETE CASCADE,

    CONSTRAINT recurring_occurrences_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT recurring_occurrences_transaction_fk
        FOREIGN KEY (transaction_id)
        REFERENCES transactions(id)
        ON DELETE SET NULL
);
```

Suggested statuses:

```text
PENDING
CONFIRMED
SKIPPED
FAILED
```

Critical uniqueness:

```sql
CREATE UNIQUE INDEX recurring_occurrence_unique_uq
ON recurring_occurrences(
    recurring_transaction_id,
    occurrence_date
);
```

This makes worker execution idempotent.

---

# 27. Budgets

Table:

```text
budgets
```

Purpose:

Stores user spending limits by category and period.

Suggested schema:

```sql
CREATE TABLE budgets (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    category_id UUID NOT NULL,

    amount BIGINT NOT NULL,

    period_type VARCHAR(20) NOT NULL DEFAULT 'MONTHLY',

    start_date DATE NOT NULL,

    end_date DATE NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    version BIGINT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT budgets_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT budgets_category_fk
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT,

    CONSTRAINT budgets_amount_positive_ck
        CHECK (amount > 0),

    CONSTRAINT budgets_date_range_ck
        CHECK (end_date >= start_date)
);
```

Initial period:

```text
MONTHLY
```

Supported status:

```text
ACTIVE
ARCHIVED
```

---

# 28. Budget Conflict Strategy

The application must prevent overlapping active budgets for the same:

```text
user
category
period
```

For a monthly MVP where periods are normalized:

```text
2026-08-01
→
2026-08-31
```

a uniqueness constraint can be used:

```sql
CREATE UNIQUE INDEX budgets_user_category_period_uq
ON budgets(user_id, category_id, start_date)
WHERE status = 'ACTIVE';
```

If arbitrary ranges are later supported, overlap detection should use stronger application/database rules.

---

# 29. Budget Derived Values

Do not persist values that can easily become inconsistent unless caching is intentionally introduced.

Values such as:

```text
spent
remaining
utilization
projected_spend
```

should normally be calculated by the finance engine.

Conceptually:

```text
spent
=
SUM(EXPENSE transactions in category and period)
```

---

# 30. Financial Goals

Table:

```text
financial_goals
```

Purpose:

Stores user-defined financial targets.

Suggested schema:

```sql
CREATE TABLE financial_goals (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    linked_account_id UUID NULL,

    name VARCHAR(150) NOT NULL,

    description TEXT NULL,

    target_amount BIGINT NOT NULL,

    current_amount BIGINT NOT NULL DEFAULT 0,

    target_date DATE NULL,

    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    version BIGINT NOT NULL DEFAULT 1,

    achieved_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT goals_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT goals_account_fk
        FOREIGN KEY (linked_account_id)
        REFERENCES accounts(id)
        ON DELETE SET NULL,

    CONSTRAINT goals_target_positive_ck
        CHECK (target_amount > 0),

    CONSTRAINT goals_current_nonnegative_ck
        CHECK (current_amount >= 0)
);
```

Status:

```text
ACTIVE
PAUSED
ACHIEVED
CANCELLED
```

Priority:

```text
LOW
MEDIUM
HIGH
```

---

# 31. Goal Current Amount

For MVP, `current_amount` may be explicitly maintained by the user.

If linked to an account later, business semantics must be clear.

Avoid assuming:

```text
entire account balance
=
goal allocation
```

unless the product explicitly defines dedicated goal accounts.

A future allocation model may use:

```text
goal_contributions
```

instead of a mutable current amount.

---

# 32. Optional Goal Contributions

If stronger goal tracking is desired, add:

```text
goal_contributions
```

Suggested schema:

```sql
CREATE TABLE goal_contributions (
    id UUID PRIMARY KEY,

    goal_id UUID NOT NULL,

    user_id UUID NOT NULL,

    amount BIGINT NOT NULL,

    contribution_date DATE NOT NULL,

    notes TEXT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT goal_contributions_goal_fk
        FOREIGN KEY (goal_id)
        REFERENCES financial_goals(id)
        ON DELETE CASCADE,

    CONSTRAINT goal_contributions_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT goal_contributions_amount_positive_ck
        CHECK (amount > 0)
);
```

Then:

```text
goal.current_amount
```

may be calculated from contributions.

For MVP, this table is optional.

---

# 33. Forecast Snapshots

Table:

```text
forecast_snapshots
```

Purpose:

Stores generated forecast results and metadata.

Forecasts are derived data, not authoritative transaction state.

Suggested schema:

```sql
CREATE TABLE forecast_snapshots (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    horizon_days INTEGER NOT NULL,

    opening_balance BIGINT NOT NULL,

    ending_balance BIGINT NOT NULL,

    minimum_balance BIGINT NOT NULL,

    projected_income BIGINT NOT NULL,

    projected_expense BIGINT NOT NULL,

    confidence VARCHAR(20) NOT NULL,

    calculation_version VARCHAR(50) NOT NULL,

    input_version VARCHAR(100) NULL,

    data_through_date DATE NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'FRESH',

    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT forecasts_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Status:

```text
FRESH
STALE
```

Confidence:

```text
LOW
MEDIUM
HIGH
```

---

# 34. Forecast Events

Table:

```text
forecast_events
```

Purpose:

Stores event-level detail for persisted forecast snapshots.

Suggested schema:

```sql
CREATE TABLE forecast_events (
    id UUID PRIMARY KEY,

    forecast_snapshot_id UUID NOT NULL,

    event_date DATE NOT NULL,

    type VARCHAR(20) NOT NULL,

    source_type VARCHAR(30) NOT NULL,

    source_id UUID NULL,

    direction VARCHAR(10) NOT NULL,

    amount BIGINT NOT NULL,

    description VARCHAR(255) NULL,

    projected_balance_after BIGINT NOT NULL,

    sequence INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT forecast_events_snapshot_fk
        FOREIGN KEY (forecast_snapshot_id)
        REFERENCES forecast_snapshots(id)
        ON DELETE CASCADE,

    CONSTRAINT forecast_events_amount_positive_ck
        CHECK (amount >= 0)
);
```

Event type:

```text
KNOWN
SCHEDULED
ESTIMATED
ASSUMED
```

Direction:

```text
IN
OUT
```

Source examples:

```text
TRANSACTION
RECURRING
USER_ASSUMPTION
HISTORICAL_ESTIMATE
```

---

# 35. Forecast Snapshot Persistence Strategy

Not every dashboard forecast request must necessarily be stored.

Possible strategy:

```text
interactive preview
→ calculate without persistence

explicit forecast generation
→ persist snapshot
```

or persist only the latest forecast.

The implementation choice should balance:

- simplicity,
- auditability,
- storage cost,
- scenario reproducibility.

For take-home scope, storing explicit generated snapshots is defensible.

---

# 36. Financial Scenarios

Table:

```text
financial_scenarios
```

Purpose:

Stores hypothetical user planning scenarios.

Suggested schema:

```sql
CREATE TABLE financial_scenarios (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    name VARCHAR(180) NOT NULL,

    description TEXT NULL,

    horizon_days INTEGER NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',

    version BIGINT NOT NULL DEFAULT 1,

    last_calculated_at TIMESTAMPTZ NULL,

    is_stale BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT scenarios_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Status:

```text
DRAFT
CALCULATED
ARCHIVED
```

---

# 37. Scenario Modifications

Table:

```text
scenario_modifications
```

Purpose:

Stores hypothetical modifications applied to a scenario.

Suggested schema:

```sql
CREATE TABLE scenario_modifications (
    id UUID PRIMARY KEY,

    scenario_id UUID NOT NULL,

    user_id UUID NOT NULL,

    type VARCHAR(40) NOT NULL,

    name VARCHAR(180) NOT NULL,

    amount BIGINT NULL,

    percentage NUMERIC(8,4) NULL,

    effective_date DATE NOT NULL,

    end_date DATE NULL,

    frequency VARCHAR(20) NULL,

    duration_months INTEGER NULL,

    source_recurring_transaction_id UUID NULL,

    category_id UUID NULL,

    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT scenario_modifications_scenario_fk
        FOREIGN KEY (scenario_id)
        REFERENCES financial_scenarios(id)
        ON DELETE CASCADE,

    CONSTRAINT scenario_modifications_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT scenario_modifications_recurring_fk
        FOREIGN KEY (source_recurring_transaction_id)
        REFERENCES recurring_transactions(id)
        ON DELETE SET NULL,

    CONSTRAINT scenario_modifications_category_fk
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE SET NULL
);
```

Types:

```text
ONE_TIME_INCOME
ONE_TIME_EXPENSE

RECURRING_INCOME
RECURRING_EXPENSE

INCOME_REDUCTION
INCOME_REMOVAL

EXPENSE_REDUCTION

SAVINGS_ADJUSTMENT
```

---

# 38. Scenario Modification Validation

Different modification types require different fields.

Examples:

```text
ONE_TIME_EXPENSE
→ amount required

INCOME_REDUCTION
→ percentage required
→ source_recurring_transaction_id usually required

INCOME_REMOVAL
→ source required

RECURRING_EXPENSE
→ amount + frequency required
```

These conditional rules belong primarily in application service validation.

---

# 39. Scenario Snapshots

Table:

```text
scenario_snapshots
```

Purpose:

Stores calculated baseline-vs-scenario results.

Suggested schema:

```sql
CREATE TABLE scenario_snapshots (
    id UUID PRIMARY KEY,

    scenario_id UUID NOT NULL,

    user_id UUID NOT NULL,

    baseline_forecast_snapshot_id UUID NULL,

    baseline_ending_balance BIGINT NOT NULL,

    scenario_ending_balance BIGINT NOT NULL,

    baseline_minimum_balance BIGINT NOT NULL,

    scenario_minimum_balance BIGINT NOT NULL,

    baseline_savings_rate NUMERIC(10,4) NULL,

    scenario_savings_rate NUMERIC(10,4) NULL,

    baseline_cash_runway_months NUMERIC(10,2) NULL,

    scenario_cash_runway_months NUMERIC(10,2) NULL,

    goal_impact JSONB NOT NULL DEFAULT '{}'::jsonb,

    assumptions JSONB NOT NULL DEFAULT '{}'::jsonb,

    calculation_version VARCHAR(50) NOT NULL,

    data_through_date DATE NOT NULL,

    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT scenario_snapshots_scenario_fk
        FOREIGN KEY (scenario_id)
        REFERENCES financial_scenarios(id)
        ON DELETE CASCADE,

    CONSTRAINT scenario_snapshots_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT scenario_snapshots_forecast_fk
        FOREIGN KEY (baseline_forecast_snapshot_id)
        REFERENCES forecast_snapshots(id)
        ON DELETE SET NULL
);
```

This snapshot allows old scenario results to remain explainable even if user data later changes.

---

# 40. Scenario Snapshot Versioning

Every calculated snapshot should include:

```text
calculation_version
data_through_date
calculated_at
```

This allows Savio to explain:

```text
This scenario was calculated
using finance-engine-v1
on 24 Aug 2026
with data available through 24 Aug 2026.
```

---

# 41. AI Insights

Table:

```text
ai_insights
```

Purpose:

Stores explainable AI outputs generated from deterministic financial signals.

Suggested schema:

```sql
CREATE TABLE ai_insights (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    type VARCHAR(40) NOT NULL,

    severity VARCHAR(20) NOT NULL,

    title VARCHAR(255) NOT NULL,

    summary TEXT NOT NULL,

    structured_context JSONB NOT NULL,

    drivers JSONB NOT NULL DEFAULT '[]'::jsonb,

    suggested_actions JSONB NOT NULL DEFAULT '[]'::jsonb,

    confidence NUMERIC(5,4) NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'NEW',

    model VARCHAR(120) NULL,

    prompt_version VARCHAR(50) NULL,

    deduplication_key VARCHAR(255) NULL,

    generated_at TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ai_insights_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

---

# 42. AI Insight Types

Initial types:

```text
SPENDING_ANOMALY
INCOME_CHANGE
BUDGET_RISK
CASHFLOW_RISK
SAVINGS_PATTERN
GOAL_RISK
RECURRING_COST
POSITIVE_TREND
```

Severity:

```text
INFO
LOW
MEDIUM
HIGH
```

Status:

```text
NEW
VIEWED
ACKNOWLEDGED
DISMISSED
```

---

# 43. AI Insight Deduplication

Recommended unique partial index:

```sql
CREATE UNIQUE INDEX ai_insights_dedup_uq
ON ai_insights(user_id, deduplication_key)
WHERE deduplication_key IS NOT NULL
  AND status <> 'DISMISSED';
```

Exact deduplication semantics may instead be application-controlled if the same insight can validly recur in different periods.

A good key may include:

```text
user
+
insight type
+
period
+
resource
```

Example:

```text
user123:budget_risk:2026-08:food
```

---

# 44. AI Insight Feedback

Table:

```text
ai_insight_feedback
```

Purpose:

Stores user feedback about AI usefulness.

Suggested schema:

```sql
CREATE TABLE ai_insight_feedback (
    id UUID PRIMARY KEY,

    insight_id UUID NOT NULL,

    user_id UUID NOT NULL,

    feedback VARCHAR(20) NOT NULL,

    comment TEXT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT insight_feedback_insight_fk
        FOREIGN KEY (insight_id)
        REFERENCES ai_insights(id)
        ON DELETE CASCADE,

    CONSTRAINT insight_feedback_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Feedback values:

```text
HELPFUL
NOT_HELPFUL
```

Potential uniqueness:

```text
one feedback per user per insight
```

---

# 45. AI Request Logs

Table:

```text
ai_requests
```

Purpose:

Provides observability for AI usage without duplicating sensitive financial data unnecessarily.

Suggested schema:

```sql
CREATE TABLE ai_requests (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    feature VARCHAR(40) NOT NULL,

    model VARCHAR(120) NULL,

    provider VARCHAR(80) NULL,

    status VARCHAR(20) NOT NULL,

    latency_ms INTEGER NULL,

    input_tokens INTEGER NULL,

    output_tokens INTEGER NULL,

    error_code VARCHAR(100) NULL,

    request_id VARCHAR(100) NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ai_requests_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Feature examples:

```text
COPILOT
INSIGHT
CATEGORY_SUGGESTION
SCENARIO_EXPLANATION
```

Status:

```text
SUCCESS
FAILED
TIMEOUT
INVALID_OUTPUT
```

Do not store full sensitive prompts in this table by default.

---

# 46. AI Copilot Conversation Persistence

Conversation persistence is implemented through:

```text
ai_conversations
ai_messages
```

and migration `000014_ai_conversations`.

However, the application should avoid storing unnecessary raw financial context in every message.

Authoritative shape:

```sql
CREATE TABLE ai_conversations (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    title VARCHAR(120) NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ai_conversations_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Messages:

```sql
CREATE TABLE ai_messages (
    id UUID PRIMARY KEY,

    conversation_id UUID NOT NULL,

    role VARCHAR(20) NOT NULL,

    content TEXT NOT NULL,
    response JSONB NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ai_messages_role_ck CHECK (role IN ('USER', 'ASSISTANT')),
    CONSTRAINT ai_messages_content_ck CHECK (char_length(content) BETWEEN 1 AND 10000),
    CONSTRAINT ai_messages_conversation_fk
        FOREIGN KEY (conversation_id)
        REFERENCES ai_conversations(id)
        ON DELETE CASCADE
);
```

Conversation access is always scoped by `workspace_id + user_id + id`.
`response` stores the validated assistant response shown at that historical
point. It must never replace current deterministic finance calculations.

---

# 47. Notifications

Table:

```text
notifications
```

Purpose:

Stores in-app notifications.

Suggested schema:

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    type VARCHAR(40) NOT NULL,

    title VARCHAR(255) NOT NULL,

    message TEXT NOT NULL,

    entity_type VARCHAR(50) NULL,

    entity_id UUID NULL,

    deduplication_key VARCHAR(255) NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'UNREAD',

    read_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT notifications_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Types:

```text
BUDGET_WARNING
BUDGET_EXCEEDED
UPCOMING_BILL
LOW_PROJECTED_BALANCE
GOAL_AT_RISK
AI_INSIGHT_AVAILABLE
SYSTEM
```

Status:

```text
UNREAD
READ
ARCHIVED
```

---

# 48. Notification Deduplication

Possible index:

```sql
CREATE UNIQUE INDEX notifications_dedup_uq
ON notifications(user_id, deduplication_key)
WHERE deduplication_key IS NOT NULL
  AND status <> 'ARCHIVED';
```

The exact lifecycle of duplicate notifications must be defined at service level.

---

# 49. Audit Logs

Table:

```text
audit_logs
```

Purpose:

Tracks important security and financial changes.

Suggested schema:

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,

    user_id UUID NULL,

    action VARCHAR(80) NOT NULL,

    entity_type VARCHAR(80) NULL,

    entity_id UUID NULL,

    request_id VARCHAR(100) NULL,

    ip_address INET NULL,

    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT audit_logs_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);
```

Indexes:

```sql
CREATE INDEX audit_logs_user_created_idx
ON audit_logs(user_id, created_at DESC);

CREATE INDEX audit_logs_entity_idx
ON audit_logs(entity_type, entity_id);
```

---

# 50. Audit Metadata Rules

Audit metadata may include:

```text
old values
new values
reason
safe identifiers
status change
```

but must not include:

```text
password
refresh token
access token
CSRF secret
API key
full sensitive AI prompt
```

---

# 51. Optional Transaction Tags

Future tables:

```text
tags
transaction_tags
```

Suggested:

```sql
CREATE TABLE tags (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    name VARCHAR(80) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tags_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Join:

```sql
CREATE TABLE transaction_tags (
    transaction_id UUID NOT NULL,
    tag_id UUID NOT NULL,

    PRIMARY KEY (transaction_id, tag_id),

    CONSTRAINT transaction_tags_transaction_fk
        FOREIGN KEY (transaction_id)
        REFERENCES transactions(id)
        ON DELETE CASCADE,

    CONSTRAINT transaction_tags_tag_fk
        FOREIGN KEY (tag_id)
        REFERENCES tags(id)
        ON DELETE CASCADE
);
```

P2 unless needed earlier.

---

# 52. Optional Import Jobs

Future CSV import tables:

```text
import_jobs
import_rows
```

Import lifecycle:

```text
UPLOADED
→ PROCESSING
→ REVIEW_REQUIRED
→ READY
→ IMPORTING
→ COMPLETED
```

Example schema:

```sql
CREATE TABLE import_jobs (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    file_key VARCHAR(500) NOT NULL,

    status VARCHAR(30) NOT NULL,

    total_rows INTEGER NOT NULL DEFAULT 0,

    valid_rows INTEGER NOT NULL DEFAULT 0,

    invalid_rows INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT import_jobs_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

This is post-MVP unless required.

---

# 53. Optional Receipt Attachments

Future receipt support may use:

```text
attachments
```

Suggested schema:

```sql
CREATE TABLE attachments (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    entity_type VARCHAR(50) NOT NULL,

    entity_id UUID NOT NULL,

    object_key VARCHAR(500) NOT NULL,

    original_filename VARCHAR(255) NOT NULL,

    mime_type VARCHAR(120) NOT NULL,

    size_bytes BIGINT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT attachments_user_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

File bytes should live in:

```text
MinIO / S3-compatible storage
```

not PostgreSQL.

---

# 54. Entity Relationship Diagram

Conceptual ERD:

```text
┌────────────────────┐
│       users        │
└─────────┬──────────┘
          │
          │ 1:N
          │
     ┌────┴─────────────────────────────────────────────┐
     │                                                  │
     ▼                                                  ▼
┌───────────────┐                               ┌─────────────────┐
│ auth_sessions │                               │  user_settings  │
└───────────────┘                               └─────────────────┘

          │
          │
          ├────────────────────────────────────────────────────┐
          │                                                    │
          ▼                                                    ▼
┌─────────────────┐                                  ┌─────────────────┐
│    accounts     │                                  │   categories    │
└───────┬─────────┘                                  └───────┬─────────┘
        │                                                    │
        │                                                    │
        ├───────────────────────┐                            │
        │                       │                            │
        ▼                       ▼                            ▼
┌─────────────────┐     ┌─────────────────┐        ┌─────────────────┐
│  transactions   │     │    transfers    │        │     budgets     │
└─────────────────┘     └─────────────────┘        └─────────────────┘
        ▲
        │
        │
┌───────┴────────────────┐
│ recurring_transactions │
└──────────┬─────────────┘
           │
           ▼
┌───────────────────────┐
│ recurring_occurrences │
└───────────────────────┘


users
  │
  ├───────────────► financial_goals
  │
  ├───────────────► forecast_snapshots
  │                       │
  │                       ▼
  │                 forecast_events
  │
  ├───────────────► financial_scenarios
  │                       │
  │                       ├──► scenario_modifications
  │                       │
  │                       └──► scenario_snapshots
  │
  ├───────────────► ai_insights
  │                       │
  │                       └──► ai_insight_feedback
  │
  ├───────────────► ai_requests
  │
  ├───────────────► notifications
  │
  └───────────────► audit_logs
```

---

# 55. Critical Relationship Rules

## User → Account

```text
1 user
→ many accounts
```

---

## User → Category

```text
1 user
→ many custom categories
```

System categories:

```text
user_id = NULL
```

---

## Account → Transaction

```text
1 account
→ many transactions
```

---

## Category → Transaction

```text
1 category
→ many transactions
```

---

## Account → Recurring Transaction

```text
1 account
→ many recurring rules
```

---

## Recurring Transaction → Occurrence

```text
1 recurring rule
→ many occurrences
```

---

## User → Scenario

```text
1 user
→ many scenarios
```

---

## Scenario → Modification

```text
1 scenario
→ many modifications
```

---

## Scenario → Snapshot

```text
1 scenario
→ many calculation snapshots
```

This allows scenario history if desired.

---

# 56. Financial Integrity Transaction Boundaries

The following operations must use database transactions.

## Create Income

```text
BEGIN

Create transaction
Update account balance
Create audit

COMMIT
```

---

## Create Expense

```text
BEGIN

Create transaction
Update account balance
Update relevant derived freshness state
Create audit

COMMIT
```

---

## Update Transaction

```text
BEGIN

Load current transaction
Check version
Reverse old financial impact
Apply new financial impact
Update transaction
Increment version
Create audit

COMMIT
```

---

## Void Transaction

```text
BEGIN

Load transaction
Reverse financial effect
Mark VOIDED (history preserved)
Create audit

COMMIT
```

---

## Transfer

```text
BEGIN

Lock / validate source
Lock / validate destination

Update source
Update destination

Create transfer
Create audit

COMMIT
```

---

## Recurring Auto Post

```text
BEGIN

Create occurrence marker
Create transaction
Update balance
Update occurrence
Advance recurring next date
Create audit

COMMIT
```

---

# 57. Balance Concurrency

Account balance changes are sensitive to concurrent writes.

Example:

```text
Balance:
Rp1,000,000

Request A:
Expense Rp100,000

Request B:
Expense Rp200,000
```

Naive behavior may produce lost updates.

The implementation should use one of:

```text
atomic SQL arithmetic
```

for example:

```sql
UPDATE accounts
SET current_balance = current_balance - ?
WHERE id = ?;
```

or explicit row locking:

```sql
SELECT ...
FOR UPDATE
```

within a transaction.

Avoid:

```text
read balance
calculate in application
write replacement value
```

without concurrency protection.

---

# 58. Optimistic Locking

Entities with user-editable mutable state should include:

```text
version BIGINT
```

Candidate resources:

```text
accounts
transactions
transfers
recurring_transactions
budgets
financial_goals
financial_scenarios
```

Update pattern:

```sql
UPDATE budgets
SET
    amount = ?,
    version = version + 1
WHERE id = ?
AND user_id = ?
AND version = ?;
```

Affected rows:

```text
1
→ success

0
→ version conflict or resource unavailable
```

Return:

```text
409 VERSION_CONFLICT
```

---

# 59. Account Deletion Policy

Hard delete allowed only if:

```text
no transactions
no transfers
no recurring rules
no important references
```

Otherwise:

```text
ARCHIVE
```

This prevents historical data loss.

---

# 60. Category Deletion Policy

Hard delete custom category allowed only if unused.

If referenced:

```text
ARCHIVED
```

Historical transactions keep the category relation.

---

# 61. Transaction Deletion Strategy

Two possible strategies:

## Option A — Soft Delete

```text
deleted_at
```

Benefits:

- simpler historical visibility.

Risks:

- all financial aggregates must consistently exclude deleted records.

## Option B — Void Status

```text
POSTED
→ VOIDED
```

with voiding metadata.

Benefits:

- clearer financial audit behavior.
- posted financial fields remain immutable.

For financial integrity, a **void-based approach** is preferable because corrections use VOID + replacement while preserving history.

The final implementation decision must remain consistent throughout analytics.

---

# 62. Recommended Transaction Voiding Extension

Potential fields:

```text
voided_at
void_reason
voided_by
```

Example:

```sql
ALTER TABLE transactions
ADD COLUMN voided_at TIMESTAMPTZ NULL,
ADD COLUMN void_reason TEXT NULL;
```

Then analytics count only:

```text
status = POSTED
```

---

# 63. Forecast Freshness Strategy

Financial changes can invalidate forecasts.

Candidate events:

```text
transaction created
transaction posted
transaction voided
transfer created
recurring rule changed
account adjustment
```

Simple implementation:

```text
UPDATE forecast_snapshots
SET status = 'STALE'
WHERE user_id = ?
AND status = 'FRESH';
```

Scenario snapshots based on the old data may also cause their scenarios to be marked:

```text
is_stale = TRUE
```

This is simple and explicit.

---

# 64. Scenario Freshness Strategy

When financial source data changes:

```text
financial_scenarios.is_stale = TRUE
```

for calculated scenarios.

Recalculation:

```text
new snapshot
→ is_stale = FALSE
```

Do not overwrite old snapshots if preserving history.

---

# 65. AI Data Boundary

The following are authoritative:

```text
accounts
transactions
transfers
recurring_transactions
budgets
financial_goals
forecast engine results
scenario engine results
```

The following are non-authoritative AI data:

```text
ai_insights
ai copilot text
category suggestions
scenario explanations
```

AI tables must not directly mutate financial state.

---

# 66. AI Structured Context

`ai_insights.structured_context` should store only the facts necessary to explain the insight.

Example:

```json
{
  "period": "2026-08",
  "category": "Food & Dining",
  "current_amount": 2400000,
  "baseline_amount": 1500000,
  "change_percent": 60,
  "primary_driver": {
    "label": "Food Delivery",
    "difference": 720000
  }
}
```

This supports explainability without requiring the model output itself to become the source of financial truth.

---

# 67. JSONB Usage Rules

JSONB should be used where:

- shape is genuinely flexible,
- data is snapshot metadata,
- fields are not heavily relational,
- fields are not core query dimensions.

Good candidates:

```text
scenario modification metadata
scenario assumptions
goal impact snapshot
AI structured context
AI drivers
AI suggested actions
audit metadata
```

Avoid placing core relational entities into large generic JSON blobs.

---

# 68. Indexing Strategy

Indexes should match expected queries.

High-priority indexes:

```text
users.email

accounts.user_id
accounts.user_id + status

transactions.user_id + date
transactions.user_id + account + date
transactions.user_id + category + date
transactions.user_id + type + date

transfers.user_id + date

recurring_transactions.user_id + status
recurring_transactions.next_occurrence_date

budgets.user_id + period
goals.user_id + status

forecast_snapshots.user_id + generated_at

financial_scenarios.user_id + updated_at

ai_insights.user_id + status + generated_at

notifications.user_id + status + created_at

audit_logs.user_id + created_at
```

---

# 69. Recurring Worker Index

Worker query:

```text
Find ACTIVE recurring rules
where next_occurrence_date <= today
```

Recommended:

```sql
CREATE INDEX recurring_due_idx
ON recurring_transactions(
    next_occurrence_date
)
WHERE status = 'ACTIVE';
```

---

# 70. AI Insight Listing Index

```sql
CREATE INDEX ai_insights_user_status_generated_idx
ON ai_insights(
    user_id,
    status,
    generated_at DESC
);
```

---

# 71. Notification Listing Index

```sql
CREATE INDEX notifications_user_status_created_idx
ON notifications(
    user_id,
    status,
    created_at DESC
);
```

---

# 72. Forecast History Index

```sql
CREATE INDEX forecast_user_generated_idx
ON forecast_snapshots(
    user_id,
    generated_at DESC
);
```

---

# 73. Scenario Listing Index

```sql
CREATE INDEX scenarios_user_updated_idx
ON financial_scenarios(
    user_id,
    updated_at DESC
);
```

---

# 74. Database Constraints vs Service Rules

Use database constraints for simple invariants:

```text
amount > 0
source != destination
unique email
foreign key integrity
valid date range
```

Use service-layer rules for contextual rules:

```text
category must match transaction type

account must belong to same user

account must be ACTIVE

budget category must be EXPENSE

scenario modification requirements depend on type

AI action requires confirmation
```

Business correctness should not depend solely on frontend validation.

---

# 75. Seed Data

Development seed should include:

```text
system categories
demo user
demo accounts
demo transactions
recurring salary
recurring rent
budget
financial goal
optional AI insight
```

Example system category seed:

```text
Income
├── Salary
├── Freelance
├── Business
├── Gift
└── Other Income

Expense
├── Food & Dining
├── Transport
├── Housing
├── Utilities
├── Shopping
├── Entertainment
├── Healthcare
├── Education
├── Subscriptions
└── Other Expense
```

Seed scripts must be safe for development and clearly separated from production data.

---

# 76. Demo Dataset

A coherent demo dataset is preferable to random records.

Example demo user:

```text
Name:
Alex

Timezone:
Asia/Jakarta

Currency:
IDR
```

Accounts:

```text
BCA Main
Rp15,000,000

GoPay
Rp750,000

Cash
Rp500,000
```

Recurring:

```text
Salary
+Rp12M
25th monthly

Rent
-Rp3M
1st monthly

Internet
-Rp450k
10th monthly

Netflix
-Rp186k
12th monthly
```

Budget:

```text
Food & Dining
Rp2M/month
```

Goal:

```text
Emergency Fund
Rp30M
```

This dataset should make the demo flow meaningful.

---

# 77. Migration Strategy

Recommended migration files:

```text
backend/migrations/

000001_create_users.up.sql
000001_create_users.down.sql

000002_create_auth_sessions.up.sql
000002_create_auth_sessions.down.sql

000003_create_user_settings.up.sql
000003_create_user_settings.down.sql

000004_create_accounts.up.sql
000004_create_accounts.down.sql

000005_create_categories.up.sql
000005_create_categories.down.sql

000006_create_transactions.up.sql
000006_create_transactions.down.sql

000007_create_transfers.up.sql
000007_create_transfers.down.sql

000008_create_recurring_transactions.up.sql
000008_create_recurring_transactions.down.sql

000009_create_recurring_occurrences.up.sql
000009_create_recurring_occurrences.down.sql

000010_create_budgets.up.sql
000010_create_budgets.down.sql

000011_create_financial_goals.up.sql
000011_create_financial_goals.down.sql

000012_create_forecast_snapshots.up.sql
000012_create_forecast_snapshots.down.sql

000013_create_forecast_events.up.sql
000013_create_forecast_events.down.sql

000014_create_financial_scenarios.up.sql
000014_create_financial_scenarios.down.sql

000015_create_scenario_modifications.up.sql
000015_create_scenario_modifications.down.sql

000016_create_scenario_snapshots.up.sql
000016_create_scenario_snapshots.down.sql

000017_create_ai_insights.up.sql
000017_create_ai_insights.down.sql

000018_create_ai_insight_feedback.up.sql
000018_create_ai_insight_feedback.down.sql

000019_create_ai_requests.up.sql
000019_create_ai_requests.down.sql

000020_create_notifications.up.sql
000020_create_notifications.down.sql

000021_create_audit_logs.up.sql
000021_create_audit_logs.down.sql
```

Migrations may be grouped differently during implementation, but dependencies must remain valid.

---

# 78. Migration Reproducibility Requirement

The database must be reproducible from:

```text
empty PostgreSQL database
```

using migration commands only.

Expected:

```text
empty DB
   ↓
migrate up
   ↓
latest schema
```

Rollback must also be tested.

At minimum:

```text
latest
→ down one
→ up one
```

should work safely.

---

# 79. Migration Dependency Order

Important dependency order:

```text
users

↓
auth_sessions
user_settings
accounts
categories

↓
transactions
transfers
recurring_transactions
budgets
financial_goals

↓
recurring_occurrences

↓
forecast

↓
scenarios

↓
AI
notifications
audit
```

Down migrations must reverse dependency order.

---

# 80. GORM Model Rules

GORM models should map explicitly to database schema.

Avoid relying on implicit behavior for critical constraints.

Example model structure:

```go
type Account struct {
    ID             uuid.UUID
    UserID         uuid.UUID
    Name           string
    Type           AccountType
    Currency       string
    InitialBalance decimal.Decimal
    CurrentBalance decimal.Decimal
    Status         AccountStatus
    Version        int64
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

Use decimal-safe library types for financial amounts.

---

# 81. Repository Query Rules

All user-owned repository queries should require ownership scope.

Prefer:

```go
FindByID(ctx, userID, accountID)
```

rather than:

```go
FindByID(ctx, accountID)
```

for user-owned data.

This makes unsafe queries harder to introduce accidentally.

---

# 82. No Unscoped Financial Repository Queries

Avoid patterns such as:

```sql
SELECT *
FROM transactions
WHERE id = $1;
```

for normal user operations.

Preferred:

```sql
SELECT *
FROM transactions
WHERE id = $1
AND user_id = $2;
```

This principle applies to:

```text
accounts
transactions
transfers
recurring
budgets
goals
forecasts
scenarios
AI insights
notifications
```

---

# 83. N+1 Prevention

List endpoints must avoid per-row queries.

Bad:

```text
Load 50 transactions

for each transaction:
    query account
    query category
```

Preferred:

```text
JOIN / preload required associations
```

or bulk-fetch related records.

Performance should remain explicit and measurable.

---

# 84. Analytics Query Strategy

Initial analytics can use PostgreSQL aggregation.

Example:

```sql
SELECT
    category_id,
    SUM(amount)
FROM transactions
WHERE user_id = $1
AND type = 'EXPENSE'
AND status = 'POSTED'
AND transaction_date BETWEEN $2 AND $3
GROUP BY category_id;
```

Do not load all transactions into Go just to calculate simple aggregate values.

---

# 85. Monthly Analytics Index Alignment

Primary analytics filters:

```text
user
transaction date
type
category
```

Indexes already defined should support this efficiently.

Query plans should be reviewed using:

```sql
EXPLAIN ANALYZE
```

for important aggregate endpoints.

---

# 86. Large Dataset Strategy

Initial pagination:

```text
offset pagination
```

is acceptable for ordinary transaction management.

Example:

```text
page
limit
```

For very large history, future optimization may move to:

```text
cursor pagination
```

especially for:

```text
transactions
audit logs
notifications
```

This is a documented trade-off rather than premature complexity.

---

# 87. Archival Strategy

Resources with historical dependencies use status rather than deletion.

Examples:

```text
Account
ACTIVE → ARCHIVED

Category
ACTIVE → ARCHIVED

Budget
ACTIVE → ARCHIVED
```

Archived records remain queryable when displaying history.

---

# 88. User Deletion

If user account deletion is implemented, the product must explicitly decide whether:

```text
hard delete all financial data
```

or:

```text
scheduled data erasure
```

For take-home MVP, user self-deletion may be out of scope.

If admin deletion exists, cascading foreign keys must be evaluated carefully.

---

# 89. Data Retention

Initial take-home scope does not require complex regulatory retention.

However:

- authentication session cleanup,
- stale AI request logs,
- expired ephemeral imports,

may have cleanup policies.

Financial transaction history should not be automatically deleted.

---

# 90. Security at Database Layer

Database credentials must:

- not be committed,
- come from environment configuration,
- use dedicated application credentials,
- avoid superuser role in normal runtime.

Production deployments should prefer TLS-enabled database connections where infrastructure supports it.

---

# 91. Sensitive Columns

Sensitive information includes:

```text
password_hash
refresh_token_hash
financial balances
transaction descriptions
AI financial context
```

These must not be unnecessarily exposed through API DTOs or logs.

---

# 92. Database Backup Consideration

For production architecture, PostgreSQL backups would be required.

This is not necessarily part of take-home implementation, but should be recognized as an operational requirement.

---

# 93. Health Check

Database health may participate in:

```text
/health
/readiness
```

Recommended distinction:

```text
liveness
→ process alive

readiness
→ application able to serve critical requests
```

PostgreSQL should generally be a required readiness dependency.

---

# 94. Redis Is Not Financial Source of Truth

If Redis is used for:

```text
queue
rate limit
short cache
```

financial state must remain in PostgreSQL.

Never rely on Redis as the authoritative source for:

```text
balances
transactions
budgets
goals
```

---

# 95. Cache Invalidation

If dashboard analytics are cached later:

```text
transaction write
transfer
adjustment
budget change
```

must invalidate relevant cached values.

Caching is P2 unless performance measurements justify it.

---

# 96. Database Testing Requirements

Integration tests should cover:

```text
migration from empty DB
rollback

foreign keys
unique constraints
check constraints

transaction creation
transaction correction
transaction voiding

transfer atomicity

recurring occurrence uniqueness

budget uniqueness

optimistic locking

ownership scoping

forecast persistence

scenario snapshot persistence
```

---

# 97. Critical Concurrency Test

Example:

```text
Initial balance:
Rp1,000,000
```

Run concurrently:

```text
Expense A:
Rp100,000

Expense B:
Rp200,000
```

Expected:

```text
Final balance:
Rp700,000
```

never:

```text
Rp800,000
or
Rp900,000
```

This validates lost-update protection.

---

# 98. Transfer Atomicity Test

Initial:

```text
Account A:
Rp1M

Account B:
Rp500k
```

Transfer:

```text
Rp300k
```

Expected:

```text
A:
Rp700k

B:
Rp800k
```

If destination update fails:

```text
A must remain Rp1M
B must remain Rp500k
transfer must not exist
```

---

# 99. Recurring Idempotency Test

Worker executes same occurrence twice:

```text
Salary
Rp12M
25 Aug
```

Expected:

```text
1 transaction only
1 occurrence only
```

Unique recurrence key ensures correctness.

---

# 100. Database MVP Scope

## P0 Tables

Must be implemented:

```text
users
auth_sessions
user_settings

accounts
categories
transactions
transfers

recurring_transactions
recurring_occurrences

budgets
financial_goals

forecast_snapshots
forecast_events

financial_scenarios
scenario_modifications
scenario_snapshots

ai_insights
ai_requests

audit_logs
```

---

## P1 Tables

High-value:

```text
ai_insight_feedback
notifications
ai_conversations
ai_messages
```

---

## P2 Tables

Optional:

```text
tags
transaction_tags

goal_contributions

attachments

import_jobs
import_rows

merchant_mappings

households
household_members
```

---

# 101. Recommended Simplification Rule

If implementation scope becomes too large, simplify by removing derived persistence before removing core business behavior.

Example:

Prefer:

```text
calculate forecast live
```

instead of removing:

```text
scenario simulator
```

Likewise, AI conversation history is less important than:

```text
structured AI tool orchestration
```

The database should support the product thesis, not maximize table count.

---

# 102. Database Source-of-Truth Hierarchy

Authoritative financial data:

```text
accounts
transactions
transfers
recurring rules
budgets
goals
```

Derived deterministic data:

```text
analytics
forecast snapshots
scenario snapshots
financial health
```

AI-generated interpretation:

```text
AI insights
AI explanations
AI conversations
```

The dependency direction is:

```text
AUTHORITATIVE FINANCIAL DATA
            ↓
DETERMINISTIC DERIVED DATA
            ↓
AI INTERPRETATION
```

It must never reverse.

---

# 103. Final Database Principle

The Savio database must make it difficult for implementation mistakes to corrupt financial history.

The intended design prioritizes:

```text
correctness
integrity
ownership
auditability
determinism
```

over clever abstractions.

The central database rule is:

> **Financial state is authoritative. Derived calculations are reproducible. AI output is advisory.**

And the overall Savio principle remains:

> **Finance Engine calculates. AI interprets. User decides.**

-- Recurring transactions: planned/scheduled financial activity.

CREATE TABLE recurring_transactions (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    category_id UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    frequency VARCHAR(20) NOT NULL,
    day_of_month INT NULL,
    day_of_week INT NULL,
    start_date DATE NOT NULL,
    end_date DATE NULL,
    description TEXT NULL,
    merchant VARCHAR(200) NULL,
    notes TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    auto_post BOOLEAN NOT NULL DEFAULT FALSE,
    version BIGINT NOT NULL DEFAULT 1,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT recurring_type_ck CHECK (type IN ('INCOME', 'EXPENSE')),
    CONSTRAINT recurring_frequency_ck CHECK (frequency IN ('DAILY', 'WEEKLY', 'MONTHLY', 'MONTH_END')),
    CONSTRAINT recurring_status_ck CHECK (status IN ('ACTIVE', 'PAUSED', 'ENDED')),
    CONSTRAINT recurring_amount_ck CHECK (amount > 0)
);
CREATE INDEX recurring_ws_status_idx ON recurring_transactions(workspace_id, status);

CREATE TABLE recurring_occurrences (
    id UUID PRIMARY KEY,
    recurring_id UUID NOT NULL REFERENCES recurring_transactions(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    due_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    posted_transaction_id UUID NULL REFERENCES transactions(id) ON DELETE SET NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT occurrences_status_ck CHECK (status IN ('PENDING', 'CONFIRMED', 'SKIPPED', 'FAILED')),
    CONSTRAINT occurrences_once_per_rule_date_uq UNIQUE (recurring_id, due_date)
);
CREATE INDEX occurrences_ws_due_idx ON recurring_occurrences(workspace_id, due_date);
CREATE INDEX occurrences_due_status_idx ON recurring_occurrences(due_date, status);
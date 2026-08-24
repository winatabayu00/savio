-- Category spending budgets.

CREATE TABLE budgets (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    version BIGINT NOT NULL DEFAULT 1,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT budgets_amount_ck CHECK (amount > 0),
    CONSTRAINT budgets_status_ck CHECK (status IN ('ACTIVE', 'CLOSED')),
    CONSTRAINT budgets_period_ck CHECK (period_end >= period_start)
);
CREATE INDEX budgets_ws_status_idx ON budgets(workspace_id, status);
CREATE INDEX budgets_ws_period_idx ON budgets(workspace_id, period_start, period_end);
-- at most one active budget per category (the app also checks period overlap)
CREATE UNIQUE INDEX budgets_ws_active_category_uq ON budgets(workspace_id, category_id) WHERE status = 'ACTIVE';
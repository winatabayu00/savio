-- Financial ledger: transactions and audit log.

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    category_id UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    transaction_date DATE NOT NULL,
    description TEXT NULL,
    merchant VARCHAR(200) NULL,
    notes TEXT NULL,
    source VARCHAR(20) NOT NULL DEFAULT 'MANUAL',
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    version BIGINT NOT NULL DEFAULT 1,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    void_reason TEXT NULL,
    voided_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    posted_at TIMESTAMPTZ NULL,
    voided_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transactions_type_ck CHECK (type IN ('INCOME', 'EXPENSE', 'ADJUSTMENT')),
    CONSTRAINT transactions_source_ck CHECK (source IN ('MANUAL', 'AI', 'IMPORT', 'RECURRING', 'SYSTEM')),
    CONSTRAINT transactions_status_ck CHECK (status IN ('DRAFT', 'POSTED', 'VOIDED')),
    -- INCOME/EXPENSE carry positive amounts (direction from type); ADJUSTMENT
    -- carries the signed balance change (+ increases, - reduces) and may be 0
    CONSTRAINT transactions_amount_ck CHECK (
        (type IN ('INCOME', 'EXPENSE') AND amount > 0)
        OR (type = 'ADJUSTMENT' AND amount <> 0)
    )
);
CREATE INDEX transactions_ws_status_date_idx ON transactions(workspace_id, status, transaction_date DESC);
CREATE INDEX transactions_ws_account_idx ON transactions(workspace_id, account_id);
CREATE INDEX transactions_ws_category_idx ON transactions(workspace_id, category_id);
CREATE INDEX transactions_ws_type_idx ON transactions(workspace_id, type, status);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    workspace_id UUID NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    actor_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NULL,
    metadata JSONB NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX audit_logs_ws_occurred_idx ON audit_logs(workspace_id, occurred_at DESC);
CREATE INDEX audit_logs_resource_idx ON audit_logs(resource_type, resource_id);
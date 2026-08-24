-- Lightweight financial goals (user-maintained progress, no reserving).

CREATE TABLE goals (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    target_amount BIGINT NOT NULL,
    current_amount BIGINT NOT NULL DEFAULT 0,
    target_date DATE NULL,
    priority VARCHAR(10) NOT NULL DEFAULT 'MEDIUM',
    linked_account_id UUID NULL REFERENCES accounts(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    version BIGINT NOT NULL DEFAULT 1,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT goals_target_ck CHECK (target_amount > 0),
    CONSTRAINT goals_current_ck CHECK (current_amount >= 0),
    CONSTRAINT goals_priority_ck CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH')),
    CONSTRAINT goals_status_ck CHECK (status IN ('ACTIVE', 'PAUSED', 'ACHIEVED', 'CANCELLED'))
);
CREATE INDEX goals_ws_status_idx ON goals(workspace_id, status);
-- Money movement between internal accounts.

CREATE TABLE account_transfers (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    from_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    to_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount BIGINT NOT NULL,
    transfer_date DATE NOT NULL,
    description TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'POSTED',
    version BIGINT NOT NULL DEFAULT 1,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    void_reason TEXT NULL,
    voided_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    voided_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT account_transfers_status_ck CHECK (status IN ('POSTED', 'VOIDED')),
    CONSTRAINT account_transfers_amount_ck CHECK (amount > 0),
    CONSTRAINT account_transfers_distinct_accounts_ck CHECK (from_account_id <> to_account_id)
);
CREATE INDEX account_transfers_ws_date_idx ON account_transfers(workspace_id, transfer_date DESC);
CREATE INDEX account_transfers_ws_from_idx ON account_transfers(workspace_id, from_account_id);
CREATE INDEX account_transfers_ws_to_idx ON account_transfers(workspace_id, to_account_id);
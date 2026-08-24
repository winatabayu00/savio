-- Notifications (background sweep output).

CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,
    title VARCHAR(160) NOT NULL,
    body TEXT NULL,
    day DATE NOT NULL DEFAULT CURRENT_DATE,
    read_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notifications_type_ck CHECK (type IN ('LOW_BALANCE', 'BUDGET_WARNING', 'BUDGET_EXCEEDED'))
);
-- dedup: one notification of a type per workspace per day (plain column,
-- avoids non-immutable expression issues on timestamptz)
CREATE UNIQUE INDEX notifications_daily_uq ON notifications(workspace_id, type, day);
CREATE INDEX notifications_user_unread_idx ON notifications(user_id) WHERE read_at IS NULL;
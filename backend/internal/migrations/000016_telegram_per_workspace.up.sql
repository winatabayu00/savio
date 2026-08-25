-- Telegram recap config becomes per-workspace. A single global bot (id=1) made
-- every other workspace hit PERMISSION_DENIED ("configured for another
-- workspace"). Now each workspace configures and owns its own bot row.

ALTER TABLE telegram_settings DROP CONSTRAINT telegram_settings_pkey;
ALTER TABLE telegram_settings DROP CONSTRAINT telegram_settings_id_check;
ALTER TABLE telegram_settings DROP COLUMN id;
DELETE FROM telegram_settings WHERE workspace_id IS NULL;
ALTER TABLE telegram_settings ALTER COLUMN workspace_id SET NOT NULL;
ALTER TABLE telegram_settings ADD PRIMARY KEY (workspace_id);
ALTER TABLE telegram_settings DROP CONSTRAINT fk_telegram_settings_workspace;
ALTER TABLE telegram_settings ADD CONSTRAINT fk_telegram_settings_workspace
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;

-- update_id is only unique per bot; with one bot per workspace the exactly-once
-- guard must be scoped per workspace or two bots' sequences would collide.
DROP TABLE telegram_processed;
CREATE TABLE telegram_processed (
    workspace_id UUID NOT NULL,
    update_id BIGINT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, update_id)
);
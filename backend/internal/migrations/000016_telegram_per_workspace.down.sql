-- ponytail: down is lossy for multiple bot rows (the old schema could only hold
-- one); it restores the singleton the pre-000016 world used. Dev rollback only.
ALTER TABLE telegram_settings DROP CONSTRAINT fk_telegram_settings_workspace;
ALTER TABLE telegram_settings DROP CONSTRAINT telegram_settings_pkey;
ALTER TABLE telegram_settings ALTER COLUMN workspace_id DROP NOT NULL;
ALTER TABLE telegram_settings ADD COLUMN id SMALLINT;

UPDATE telegram_settings SET id = 1;
INSERT INTO telegram_settings (id) VALUES (1) ON CONFLICT DO NOTHING;

ALTER TABLE telegram_settings ADD CONSTRAINT telegram_settings_pkey PRIMARY KEY (id);
ALTER TABLE telegram_settings ADD CONSTRAINT telegram_settings_id_check CHECK (id = 1);
ALTER TABLE telegram_settings ADD CONSTRAINT fk_telegram_settings_workspace
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE SET NULL;

DROP TABLE telegram_processed;
CREATE TABLE telegram_processed (
    update_id BIGINT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
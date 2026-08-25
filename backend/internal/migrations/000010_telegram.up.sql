-- Telegram recap bot configuration + processed-update dedup guard.
-- Singleton row id = 1. A single bot is bound to one workspace (the OWNER who
-- configured it); telegram_processed makes the long-poll exactly-once even if
-- the worker crashes between transaction creation and offset ack.

CREATE TABLE telegram_settings (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    bot_token TEXT NOT NULL DEFAULT '',
    chat_id TEXT NOT NULL DEFAULT '',
    workspace_id UUID,
    last_update_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_telegram_settings_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id) ON DELETE SET NULL
);

INSERT INTO telegram_settings (id) VALUES (1)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE telegram_processed (
    update_id BIGINT PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
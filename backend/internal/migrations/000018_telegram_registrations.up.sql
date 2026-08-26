CREATE TABLE telegram_registration_codes (
    code_hash TEXT PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE telegram_chat_bindings (
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    telegram_user_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (workspace_id, telegram_user_id, chat_id),
    UNIQUE (telegram_user_id, chat_id)
);

CREATE UNIQUE INDEX telegram_settings_bot_token_unique
    ON telegram_settings (bot_token)
    WHERE bot_token <> '';

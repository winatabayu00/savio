-- Runtime AI provider configuration. Replaces AI_* env vars for already-running installs.
-- Singleton row id = 1. Started with provider='pending' so the application can seed
-- it from environment defaults on first startup; user edits persist across restarts.

CREATE TABLE ai_settings (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    provider VARCHAR(30) NOT NULL DEFAULT 'mock',
    base_url TEXT,
    api_key TEXT,
    model VARCHAR(120) NOT NULL DEFAULT 'gpt-4o-mini',
    timeout_seconds INT NOT NULL DEFAULT 20,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO ai_settings (id, enabled, provider, base_url, api_key, model, timeout_seconds)
VALUES (1, FALSE, 'pending', NULL, NULL, 'gpt-4o-mini', 20);
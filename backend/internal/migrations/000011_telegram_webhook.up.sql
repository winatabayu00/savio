-- Webhook mode: Telegram pushes updates to a public HTTPS URL instead of the
-- worker long-polling. webhook_secret is a random path component guarding the
-- unauthenticated webhook endpoint; webhook_url is the registered public base.

ALTER TABLE telegram_settings
    ADD COLUMN webhook_secret TEXT NOT NULL DEFAULT '',
    ADD COLUMN webhook_url TEXT NOT NULL DEFAULT '';
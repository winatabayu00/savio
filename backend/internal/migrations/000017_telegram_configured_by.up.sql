-- Telegram-driven transactions must be attributed to the user who configured
-- the bot (the one logged in when settings/webhook were saved), not an
-- arbitrary first OWNER.
ALTER TABLE telegram_settings
    ADD COLUMN configured_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

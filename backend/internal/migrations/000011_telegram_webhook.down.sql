ALTER TABLE telegram_settings
    DROP COLUMN IF EXISTS webhook_secret,
    DROP COLUMN IF EXISTS webhook_url;
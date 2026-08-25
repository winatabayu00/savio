ALTER TABLE audit_logs
    ADD COLUMN actor_type VARCHAR(20) NOT NULL DEFAULT 'USER',
    ADD COLUMN reason TEXT NULL,
    ADD COLUMN before_data JSONB NULL,
    ADD COLUMN after_data JSONB NULL;

ALTER TABLE audit_logs
    ADD CONSTRAINT audit_logs_actor_type_check CHECK (actor_type IN ('USER', 'AI', 'SYSTEM'));

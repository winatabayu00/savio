ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_actor_type_check;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS after_data;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS before_data;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS reason;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS actor_type;

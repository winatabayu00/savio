ALTER TABLE transactions
    DROP CONSTRAINT transactions_source_ck,
    ADD CONSTRAINT transactions_source_ck CHECK (source IN ('MANUAL', 'AI', 'IMPORT', 'RECURRING', 'SYSTEM'));
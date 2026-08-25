-- Selectable AI personality driving Copilot + Insight tone. 'balanced' is the
-- default Savio Copilot; 'lenna' is the personal financial advisor persona.
ALTER TABLE ai_settings ADD COLUMN persona varchar(20) NOT NULL DEFAULT 'balanced';
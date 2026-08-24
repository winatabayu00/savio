-- Scenario simulator: non-destructive overlay state only.

CREATE TABLE scenarios (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    description TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    is_stale BOOLEAN NOT NULL DEFAULT FALSE,
    baseline_snapshot TEXT NULL,
    result TEXT NULL,
    calculation_version VARCHAR(10) NULL,
    version BIGINT NOT NULL DEFAULT 1,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT scenarios_status_ck CHECK (status IN ('DRAFT', 'CALCULATED'))
);
CREATE INDEX scenarios_ws_idx ON scenarios(workspace_id, created_at DESC);

CREATE TABLE scenario_modifications (
    id UUID PRIMARY KEY,
    scenario_id UUID NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,
    amount BIGINT NOT NULL,
    frequency VARCHAR(20) NULL,
    narrative TEXT NULL,
    position INT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT mods_type_ck CHECK (type IN (
        'ONE_TIME_EXPENSE','ONE_TIME_INCOME','RECURRING_EXPENSE','RECURRING_INCOME',
        'INCOME_REDUCTION','INCOME_REMOVAL','EXPENSE_REDUCTION'))
);
CREATE INDEX mods_scenario_idx ON scenario_modifications(scenario_id);
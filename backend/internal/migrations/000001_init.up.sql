-- Initial schema: identity, workspaces, sessions, settings, accounts, categories.

CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    timezone VARCHAR(100) NOT NULL DEFAULT 'Asia/Jakarta',
    default_currency CHAR(3) NOT NULL DEFAULT 'IDR',
    locale VARCHAR(20) NOT NULL DEFAULT 'id-ID',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    email_verified_at TIMESTAMPTZ NULL,
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_status_ck CHECK (status IN ('ACTIVE', 'DISABLED'))
);

CREATE UNIQUE INDEX users_email_lower_uq ON users (LOWER(email));

CREATE TABLE workspaces (
    id UUID PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    base_currency CHAR(3) NOT NULL DEFAULT 'IDR',
    timezone VARCHAR(100) NOT NULL DEFAULT 'Asia/Jakarta',
    created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE workspace_memberships (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'MEMBER',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT workspace_memberships_role_ck CHECK (role IN ('OWNER', 'MEMBER', 'VIEWER')),
    CONSTRAINT workspace_memberships_status_ck CHECK (status IN ('ACTIVE', 'REMOVED'))
);

CREATE UNIQUE INDEX workspace_memberships_ws_user_uq
    ON workspace_memberships(workspace_id, user_id);
CREATE INDEX workspace_memberships_user_ws_idx ON workspace_memberships(user_id, workspace_id);
CREATE INDEX workspace_memberships_ws_role_idx ON workspace_memberships(workspace_id, role);

CREATE TABLE auth_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    user_agent TEXT NULL,
    ip_address INET NULL,
    device_name VARCHAR(150) NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX auth_sessions_user_id_idx ON auth_sessions(user_id);
CREATE INDEX auth_sessions_expires_at_idx ON auth_sessions(expires_at);

CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    ai_insights_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ai_copilot_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    budget_warning_threshold NUMERIC(5,2) NOT NULL DEFAULT 80.00,
    low_balance_threshold BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    type VARCHAR(30) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'IDR',
    opening_balance BIGINT NOT NULL DEFAULT 0,
    institution_name VARCHAR(150) NULL,
    description TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    version BIGINT NOT NULL DEFAULT 1,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    archived_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT accounts_type_ck CHECK (type IN ('CASH', 'BANK', 'EWALLET', 'SAVINGS', 'OTHER')),
    CONSTRAINT accounts_status_ck CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    CONSTRAINT accounts_balance_nonnegative_ck CHECK (opening_balance >= 0)
);
CREATE INDEX accounts_ws_status_idx ON accounts(workspace_id, status);
CREATE UNIQUE INDEX accounts_ws_name_active_uq ON accounts(workspace_id, LOWER(name)) WHERE status = 'ACTIVE';

CREATE TABLE categories (
    id UUID PRIMARY KEY,
    workspace_id UUID NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    type VARCHAR(20) NOT NULL,
    parent_id UUID NULL REFERENCES categories(id) ON DELETE SET NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    icon VARCHAR(100) NULL,
    description TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT categories_type_ck CHECK (type IN ('INCOME', 'EXPENSE')),
    CONSTRAINT categories_status_ck CHECK (status IN ('ACTIVE', 'ARCHIVED')),
    CONSTRAINT categories_scope_ck CHECK (
        (is_system = TRUE AND workspace_id IS NULL) OR (is_system = FALSE AND workspace_id IS NOT NULL)
    )
);
CREATE UNIQUE INDEX categories_system_name_type_uq ON categories(LOWER(name), type) WHERE is_system = TRUE;
CREATE UNIQUE INDEX categories_ws_name_type_uq ON categories(workspace_id, LOWER(name), type) WHERE workspace_id IS NOT NULL;
CREATE INDEX categories_ws_type_idx ON categories(workspace_id, type, status);
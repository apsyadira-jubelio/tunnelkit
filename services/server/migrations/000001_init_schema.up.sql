-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'member',
    created_at  TIMESTAMPTZ DEFAULT now()
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);

-- API Keys table
CREATE TABLE api_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    key_hash    TEXT NOT NULL UNIQUE,
    key_prefix  TEXT NOT NULL,
    scopes      TEXT[] NOT NULL DEFAULT '{}',
    last_used   TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    revoked     BOOLEAN DEFAULT false,
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- Tunnels table
CREATE TABLE tunnels (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id),
    name        TEXT NOT NULL,
    protocol    TEXT NOT NULL,
    subdomain   TEXT UNIQUE,
    remote_port INT,
    auth_type   TEXT DEFAULT 'none',
    auth_config JSONB,
    ip_allowlist CIDR[],
    status      TEXT NOT NULL DEFAULT 'inactive',
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_tunnels_user_id ON tunnels(user_id);
CREATE INDEX idx_tunnels_subdomain ON tunnels(subdomain);
CREATE INDEX idx_tunnels_status ON tunnels(status);

-- Audit Logs table
CREATE TABLE audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    actor_id    UUID REFERENCES users(id),
    action      TEXT NOT NULL,
    resource    TEXT,
    ip_address  INET,
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- +migrate Down

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS tunnels;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "pgcrypto";

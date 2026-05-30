-- +migrate Up

-- Tunnel Logs table
CREATE TABLE tunnel_logs (
    id              BIGSERIAL PRIMARY KEY,
    tunnel_id       UUID REFERENCES tunnels(id) ON DELETE CASCADE,
    method          TEXT NOT NULL,
    path            TEXT NOT NULL,
    query           TEXT,
    status_code     INT NOT NULL,
    duration_ms     INT NOT NULL DEFAULT 0,
    request_body    BYTEA,
    response_body   BYTEA,
    request_headers  JSONB DEFAULT '{}',
    response_headers JSONB DEFAULT '{}',
    client_ip       INET NOT NULL,
    user_agent      TEXT,
    error           TEXT,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_tunnel_logs_tunnel_id ON tunnel_logs(tunnel_id);
CREATE INDEX idx_tunnel_logs_created_at ON tunnel_logs(created_at);
CREATE INDEX idx_tunnel_logs_tunnel_created ON tunnel_logs(tunnel_id, created_at DESC);

-- +migrate Down

DROP TABLE IF EXISTS tunnel_logs;

CREATE TABLE command_runs (
    id BIGSERIAL PRIMARY KEY,
    command TEXT NOT NULL,
    profile TEXT NOT NULL,
    status TEXT NOT NULL,
    output TEXT,
    error_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE inspect_snapshots (
    id BIGSERIAL PRIMARY KEY,
    app_name TEXT NOT NULL,
    profile TEXT NOT NULL,
    payload_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_command_runs_created_at ON command_runs (created_at DESC);
CREATE INDEX idx_inspect_snapshots_app ON inspect_snapshots (app_name, created_at DESC);
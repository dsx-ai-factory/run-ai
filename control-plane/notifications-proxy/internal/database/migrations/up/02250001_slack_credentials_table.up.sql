CREATE TABLE IF NOT EXISTS slack_credentials (
    id SMALLINT PRIMARY KEY CHECK (id = 1), -- only one row allowed, fixed id lets us safely upsert by ON CONFLICT
    app_id TEXT,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

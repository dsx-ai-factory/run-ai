CREATE TABLE IF NOT EXISTS application_chart_uploads (
    id          UUID PRIMARY KEY,
    tenant_id   VARCHAR NOT NULL,
    chart_bytes BYTEA NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS application_chart_uploads_created_at
    ON application_chart_uploads (created_at);

CREATE TABLE IF NOT EXISTS cluster_histories (
    id               SERIAL,
    tenant_id        INTEGER NOT NULL,
    cluster_id       UUID,
    version          TEXT,
    update_timestamp TIMESTAMPTZ,
    created_at       TIMESTAMPTZ DEFAULT now(),
    updated_at       TIMESTAMPTZ DEFAULT now(),
    deleted_at       TIMESTAMPTZ,

    PRIMARY KEY (cluster_id, version),
    FOREIGN KEY (cluster_id) REFERENCES clusters(id)
);
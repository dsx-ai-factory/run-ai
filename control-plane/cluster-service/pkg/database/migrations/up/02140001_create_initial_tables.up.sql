CREATE TABLE IF NOT EXISTS node (
    id                  UUID NOT NULL PRIMARY KEY,
    name                TEXT NOT NULL,
    status              TEXT NOT NULL,
    conditions          JSONB,
    taints              JSONB,
    node_pool           TEXT NOT NULL,
    gpu_type            TEXT NOT NULL,
    gpu_count           INT NOT NULL,
    cluster_uuid        TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ,

    CONSTRAINT unique_node_name_per_cluster UNIQUE (name, cluster_uuid)
);

CREATE INDEX idx_cluster ON node (cluster_uuid);

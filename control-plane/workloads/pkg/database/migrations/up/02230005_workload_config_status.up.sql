DROP TABLE IF EXISTS workload_type_statuses;
CREATE TABLE IF NOT EXISTS workload_type_statuses (
    id               UUID NOT NULL PRIMARY KEY,
    workload_type_id UUID NOT NULL,
    cluster_id       VARCHAR NOT NULL,
    tenant_id        VARCHAR NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    phase            VARCHAR NOT NULL,
    CONSTRAINT workload_type_id_status_fk FOREIGN KEY (workload_type_id) REFERENCES workload_types(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_status_cluster ON workload_type_statuses(workload_type_id,cluster_id,tenant_id);

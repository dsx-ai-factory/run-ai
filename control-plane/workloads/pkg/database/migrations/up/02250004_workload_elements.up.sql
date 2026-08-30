-- workload_elements: stores hierarchy levels 1-3 (root synthesized at query time)
CREATE TABLE IF NOT EXISTS workload_elements (
    id              VARCHAR NOT NULL,
    workload_id     UUID NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    tenant_id       BIGINT NOT NULL,
    cluster_id      UUID NOT NULL,
    name            VARCHAR NOT NULL,
    level           INT NOT NULL,
    parent_id       VARCHAR,
    instance_key    VARCHAR,
    replica_key     VARCHAR,
    k8s_ref_uid     VARCHAR,
    k8s_ref_name    VARCHAR,
    gvk_group       VARCHAR,
    gvk_version     VARCHAR,
    gvk_kind        VARCHAR,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT workload_elements_pk PRIMARY KEY (workload_id, id)
);

-- Total element count on workloads (for truncation awareness)
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS structure_elements_total_count INT;

-- pod_elements: join table linking pods to structure elements by hierarchy level
CREATE TABLE IF NOT EXISTS pod_elements (
    pod_id      UUID NOT NULL REFERENCES pods(id) ON DELETE CASCADE,
    element_id  VARCHAR NOT NULL,
    level       INT NOT NULL,
    CONSTRAINT pod_elements_pk PRIMARY KEY (pod_id, level)
);
CREATE INDEX IF NOT EXISTS idx_pod_elements_element_level ON pod_elements (element_id, level);

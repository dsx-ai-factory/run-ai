ALTER TABLE node
    ADD COLUMN IF NOT EXISTS gpu_memory_bytes BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_node_tenant_cluster_node_pool ON node (tenant_id, cluster_uuid, node_pool);

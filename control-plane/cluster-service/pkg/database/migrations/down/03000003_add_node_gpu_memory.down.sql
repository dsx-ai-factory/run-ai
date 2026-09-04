DROP INDEX IF EXISTS idx_node_tenant_cluster_node_pool;

ALTER TABLE node
    DROP COLUMN IF EXISTS gpu_memory_bytes;

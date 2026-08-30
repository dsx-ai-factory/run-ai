DROP INDEX IF EXISTS nodes_search_vector_idx;
ALTER TABLE IF EXISTS node DROP COLUMN IF EXISTS search_vector;

DROP INDEX IF EXISTS idx_node_nv_link_clique_id;
DROP INDEX IF EXISTS idx_node_nv_link_domain_uid;
DROP INDEX IF EXISTS idx_node_gpu_type;
DROP INDEX IF EXISTS idx_node_node_pool;
DROP INDEX IF EXISTS idx_node_status;
DROP INDEX IF EXISTS idx_node_name;

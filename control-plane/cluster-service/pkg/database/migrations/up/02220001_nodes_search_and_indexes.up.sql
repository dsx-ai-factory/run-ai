-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE INDEX IF NOT EXISTS idx_node_name ON node (name);
CREATE INDEX IF NOT EXISTS idx_node_status ON node (status);
CREATE INDEX IF NOT EXISTS idx_node_node_pool ON node (node_pool);
CREATE INDEX IF NOT EXISTS idx_node_gpu_type ON node (gpu_type);
CREATE INDEX IF NOT EXISTS idx_node_nv_link_domain_uid ON node (nv_link_domain_uid);
CREATE INDEX IF NOT EXISTS idx_node_nv_link_clique_id ON node (nv_link_clique_id);

ALTER TABLE IF EXISTS node ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS
    (to_tsvector(
            'english',
            COALESCE(name::text, '') || ' ' ||
            COALESCE(status::text, '') || ' ' ||
            COALESCE(node_pool::text, '') || ' ' ||
            COALESCE(gpu_type::text, '') || ' ' ||
            COALESCE(nv_link_domain_uid::text, '') || ' ' ||
            COALESCE(nv_link_clique_id::text, ''))) STORED;
CREATE INDEX IF NOT EXISTS nodes_search_vector_idx ON node using gin(search_vector);

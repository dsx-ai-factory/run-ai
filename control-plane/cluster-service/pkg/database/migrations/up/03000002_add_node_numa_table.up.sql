-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS node_numa
(
    id                           UUID PRIMARY KEY,
    tenant_id                    INT            NOT NULL,
    cluster_id                   UUID           NOT NULL,
    node_id                      UUID           NOT NULL,
    name                       TEXT           NOT NULL,
    total_cpu_cores            DECIMAL(12, 3) NOT NULL,
    allocated_cpu_cores        DECIMAL(12, 3) NOT NULL,
    total_cpu_memory_bytes     BIGINT         NOT NULL,
    allocated_cpu_memory_bytes BIGINT         NOT NULL,
    total_gpus                 INT            NOT NULL,
    allocated_gpus             INT            NOT NULL,
    created_at                 TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at                 TIMESTAMPTZ    NOT NULL DEFAULT now(),

    CONSTRAINT node_numa_unique_name_per_node UNIQUE (node_id, name),
    CONSTRAINT node_numa_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES clusters (id) ON DELETE CASCADE,
    CONSTRAINT node_numa_node_id_fk FOREIGN KEY (node_id) REFERENCES node (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_node_numa_cluster_tenant ON node_numa (cluster_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_node_numa_node_id ON node_numa (node_id);

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools ADD COLUMN network_topology_name VARCHAR(255) NULL DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN network_topology_id VARCHAR(255) NULL DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_network_topology_name ON nodepools (tenant_id, cluster_id, network_topology_name);

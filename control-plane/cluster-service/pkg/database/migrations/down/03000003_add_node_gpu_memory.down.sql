-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_node_tenant_cluster_node_pool;

ALTER TABLE node
    DROP COLUMN IF EXISTS gpu_memory_bytes;

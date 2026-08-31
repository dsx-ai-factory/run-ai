-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_nt_tenant_cluster_id;
DROP INDEX IF EXISTS idx_nt_source;
DROP INDEX IF EXISTS nt_unique_name_per_tenant_cluster;

DROP TABLE IF EXISTS network_topologies;


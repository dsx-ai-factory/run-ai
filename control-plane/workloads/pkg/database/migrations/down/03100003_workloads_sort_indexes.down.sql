-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_workloads_tenant_cluster_phase;
DROP INDEX IF EXISTS idx_workloads_tenant_cluster_created;
DROP INDEX IF EXISTS idx_workloads_tenant_cluster_name;

DROP INDEX IF EXISTS idx_view_workloads_partial_tenant_cluster_phase;
DROP INDEX IF EXISTS idx_view_workloads_partial_tenant_cluster_created;
DROP INDEX IF EXISTS idx_view_workloads_partial_tenant_cluster_name;

DROP INDEX IF EXISTS idx_view_workloads_tenant_cluster_phase;
DROP INDEX IF EXISTS idx_view_workloads_tenant_cluster_created;
DROP INDEX IF EXISTS idx_view_workloads_tenant_cluster_name;

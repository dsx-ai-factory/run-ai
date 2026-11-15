-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Same indexes as 02260008, which databases upgraded from v2.26 already have: 03000007 drops and
-- recreates both materialized views, so their indexes are recreated here; the table indexes are
-- created here for databases that never ran 02260008.

-- #################################################################
--    Sort indexes for GET /api/v1/workloads?sortBy=phase|createdAt|name
-- #################################################################
-- The API orders by "<col> ASC NULLS FIRST" or "<col> DESC NULLS LAST", then "id DESC", which no
-- default-order index can serve. Each index below is declared DESC NULLS LAST: read forward it serves
-- sortOrder=desc, read backward sortOrder=asc. The (tenant_id, cluster_id, phase) prefix also serves
-- the pending-time telemetry. The name index also serves name lookups (CLI describe, attach, logs, ...).
-- Created on the table (live view) and on both materialized views (read when
-- view_refresh_rate_milli_seconds > 0). A migration that recreates the materialized views must
-- recreate their indexes as well.

CREATE INDEX IF NOT EXISTS idx_workloads_tenant_cluster_phase
    ON workloads (tenant_id, cluster_id, phase DESC NULLS LAST, phase_updated_at DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_workloads_tenant_cluster_created
    ON workloads (tenant_id, cluster_id, created_at DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_workloads_tenant_cluster_name
    ON workloads (tenant_id, cluster_id, name DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_view_workloads_partial_tenant_cluster_phase
    ON workloads_partial_materialized_view (tenant_id, cluster_id, phase DESC NULLS LAST, phase_updated_at DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_view_workloads_partial_tenant_cluster_created
    ON workloads_partial_materialized_view (tenant_id, cluster_id, created_at DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_view_workloads_partial_tenant_cluster_name
    ON workloads_partial_materialized_view (tenant_id, cluster_id, name DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_view_workloads_tenant_cluster_phase
    ON workloads_materialized_view (tenant_id, cluster_id, phase DESC NULLS LAST, phase_updated_at DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_view_workloads_tenant_cluster_created
    ON workloads_materialized_view (tenant_id, cluster_id, created_at DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_view_workloads_tenant_cluster_name
    ON workloads_materialized_view (tenant_id, cluster_id, name DESC NULLS LAST, id DESC)
    WHERE deleted_at IS NULL;

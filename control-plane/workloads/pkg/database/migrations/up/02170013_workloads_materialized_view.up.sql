-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create MATERIALIZED VIEW IF NOT EXISTS workloads_materialized_view AS (SELECT * from workloads_view);

CREATE UNIQUE INDEX IF NOT EXISTS idx_view_workloads_id ON workloads_materialized_view (id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_type ON workloads_materialized_view (type);
CREATE INDEX IF NOT EXISTS idx_view_workloads_name ON workloads_materialized_view (name);
CREATE INDEX IF NOT EXISTS idx_view_workloads_project_name ON workloads_materialized_view (project_name);
CREATE INDEX IF NOT EXISTS idx_view_workloads_creation_time ON workloads_materialized_view (created_at);
CREATE INDEX IF NOT EXISTS idx_view_workloads_deletion_time ON workloads_materialized_view (deleted_at) where deleted_at is not null;
CREATE INDEX IF NOT EXISTS idx_view_workloads_submitted_by ON workloads_materialized_view (submitted_by) where submitted_by <> '';
CREATE INDEX IF NOT EXISTS idx_view_workloads_tenant ON workloads_materialized_view (tenant_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_cluster ON workloads_materialized_view (cluster_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_project ON workloads_materialized_view (project_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_department_id ON workloads_materialized_view (department_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_department_name ON workloads_materialized_view (department_name);

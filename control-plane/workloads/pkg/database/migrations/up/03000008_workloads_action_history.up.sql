-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS workloads_action_history (
    id uuid PRIMARY KEY,
    tenant_id bigint NOT NULL,
    cluster_id uuid NOT NULL,
    workload_id uuid NOT NULL,
    workload_name varchar NOT NULL,
    project_id varchar NOT NULL,
    department_id varchar NOT NULL,
    subject varchar NOT NULL,
    subject_type varchar NOT NULL,
    action varchar NOT NULL CONSTRAINT workloads_action_history_action_check
        CHECK (action IN ('create', 'update', 'delete', 'suspend', 'resume')),
    accepted_at timestamp with time zone NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_workloads_action_history_tenant_workload_accepted_at
    ON workloads_action_history (tenant_id, workload_id, accepted_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_workloads_action_history_tenant_cluster
    ON workloads_action_history (tenant_id, cluster_id);

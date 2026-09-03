-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS external_workloads_type_config_view;

UPDATE external_workloads_type_config
SET updated_by = NULL
WHERE updated_by = 'Run.ai';
UPDATE external_workloads_type_config
SET created_by = NULL
WHERE created_by = 'Run.ai';
UPDATE workload_types
SET created_by = NULL
WHERE created_by = 'Run.ai';

ALTER TABLE workload_types ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE workload_types ADD COLUMN IF NOT EXISTS updated_by VARCHAR;
ALTER TABLE workload_types RENAME COLUMN source TO type;

CREATE VIEW external_workloads_type_config_view AS
SELECT
    workload_types.id,
    workload_types.name,
    workload_types.created_at,
    workload_types.created_by,
    workload_types.updated_at,
    workload_types.updated_by,
    COALESCE(category_tenant.id, category_system.id) AS category_id,
    COALESCE(priority_tenant.id, priority_system.id) AS priority_id,
    COALESCE(category_tenant.name, category_system.name) AS category_name,
    COALESCE(priority_tenant.name, priority_system.name) AS priority_name,
    COALESCE(config_tenant.scope_type, config_system.scope_type) AS scope_type,
    COALESCE(config_tenant.tenant_id, config_system.tenant_id) AS tenant_id
FROM workload_types
         JOIN external_workloads_type_config config_tenant
              ON config_tenant.workload_type_id = workload_types.id
         JOIN external_workloads_type_config config_system
              ON config_system.workload_type_id = workload_types.id AND config_system.scope_type = 'system'
         LEFT JOIN workloads_categories category_tenant
                   ON category_tenant.id = config_tenant.category_id
         LEFT JOIN priorities priority_tenant
                   ON priority_tenant.id = config_tenant.priority_id
         LEFT JOIN workloads_categories category_system
                   ON category_system.id = config_system.category_id
         LEFT JOIN priorities priority_system
                   ON priority_system.id = config_system.priority_id;
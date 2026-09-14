-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS external_workloads_type_config_view;

CREATE VIEW external_workloads_type_config_view AS
SELECT
    workload_types.id,
    workload_types.name,
    workload_types.group,
    workload_types.source,
    config.created_at,
    config.created_by,
    config.updated_at,
    config.updated_by,
    COALESCE(
            (SELECT array_to_json(array_agg(row_to_json(rd) ORDER BY rd.created_at))
             FROM workload_type_resource_interfaces rd
             WHERE rd.workload_type_id = workload_types.id
               AND (
                 -- First priority: match config's scope
                 (rd.scope_type = config.scope_type AND rd.tenant_id IS NOT DISTINCT FROM config.tenant_id)
                     OR
                     -- Fallback: system scope, only if no config-scope RIs exist
                 (rd.scope_type = 'system' AND rd.tenant_id IS NULL
                     AND NOT EXISTS (
                         SELECT 1 FROM workload_type_resource_interfaces ri
                         WHERE ri.workload_type_id = workload_types.id
                           AND ri.scope_type = config.scope_type
                           AND ri.tenant_id IS NOT DISTINCT FROM config.tenant_id
                     ))
                 )),
            '[]'::json
    ) AS resource_interfaces,
    COALESCE(category.id, system_category.id) AS category_id,
    COALESCE(priority.id, system_priority.id) AS priority_id,
    COALESCE(category.name, system_category.name) AS category_name,
    COALESCE(priority.name, system_priority.name) AS priority_name,
    config.scope_type,
    config.tenant_id
FROM workload_types
         INNER JOIN external_workloads_type_config config
                    ON config.workload_type_id = workload_types.id
         LEFT JOIN external_workloads_type_config config_system
                   ON config_system.workload_type_id = workload_types.id AND config_system.scope_type = 'system'

         LEFT JOIN workloads_categories category
                   ON category.id = config.category_id
         LEFT JOIN priorities priority
                   ON priority.id = config.priority_id
         LEFT JOIN workloads_categories system_category
                   ON system_category.id = config_system.category_id
         LEFT JOIN priorities system_priority
                   ON system_priority.id = config_system.priority_id;

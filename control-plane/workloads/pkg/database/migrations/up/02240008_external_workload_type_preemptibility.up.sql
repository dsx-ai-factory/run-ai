-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS external_workloads_type_config_view;

ALTER TABLE external_workloads_type_config ADD COLUMN IF NOT EXISTS preemptibility VARCHAR(50);

-- Initialize preemptibility for existing rows based on category
UPDATE external_workloads_type_config AS ewtc
SET preemptibility = 'preemptible'
FROM workloads_categories AS wc
WHERE ewtc.tenant_id IS NULL
  AND ewtc.scope_type = 'system'
  AND ewtc.category_id = wc.id
  AND wc.name = 'Train';

UPDATE external_workloads_type_config AS ewtc
SET preemptibility = 'non-preemptible'
FROM workloads_categories AS wc
WHERE ewtc.tenant_id IS NULL
  AND ewtc.scope_type = 'system'
  AND ewtc.category_id = wc.id
  AND wc.name IN ('Build', 'Deploy');

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
    COALESCE((
                 SELECT array_to_json(array_agg(row_to_json(rd) ORDER BY rd.created_at))
                 FROM workload_type_resource_interfaces rd
                 WHERE rd.workload_type_id = workload_types.id
                   AND rd.scope_type = config.scope_type
                   AND ((rd.tenant_id IS NULL AND config.tenant_id IS NULL) OR rd.tenant_id = config.tenant_id)
             ), '[]'::json) AS resource_interfaces,
    COALESCE(category.id, system_category.id) AS category_id,
    COALESCE(priority.id, system_priority.id) AS priority_id,
    COALESCE(category.name, system_category.name) AS category_name,
    COALESCE(priority.name, system_priority.name) AS priority_name,
    COALESCE(config.preemptibility, config_system.preemptibility) AS preemptibility,
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
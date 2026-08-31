-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS resources_aggregated_view;
CREATE VIEW resources_aggregated_view AS
(
SELECT org_unit_data_id,
       MAX(resources.updated_at) AS resources_updated_at,
       json_agg(resources.* ORDER BY id) AS resources_json,
       CASE
           WHEN SUM(CASE WHEN resources.gpu_deserved IS NULL THEN 1 ELSE 0 END) > 0 THEN NULL
           WHEN SUM(CASE WHEN resources.gpu_deserved = -1 THEN 1 ELSE 0 END) > 0 THEN -1
           ELSE SUM(resources.gpu_deserved)
           END AS total_gpu_quota,
       CASE
           WHEN SUM(CASE WHEN resources.cpu_deserved IS NULL THEN 1 ELSE 0 END) > 0 THEN NULL
           WHEN SUM(CASE WHEN resources.cpu_deserved = -1 THEN 1 ELSE 0 END) > 0 THEN -1
           ELSE SUM(resources.cpu_deserved)
           END AS total_cpu_quota,
       CASE
           WHEN SUM(CASE WHEN resources.memory_deserved IS NULL THEN 1 ELSE 0 END) > 0 THEN NULL
           WHEN SUM(CASE WHEN resources.memory_deserved = -1 THEN 1 ELSE 0 END) > 0 THEN -1
           ELSE ROUND(SUM(
                              CASE
                                  WHEN resources.memory_units IS NULL THEN resources.memory_deserved * 1.048576
                                  WHEN resources.memory_units = 'Mib' THEN resources.memory_deserved * 1.048576
                                  WHEN resources.memory_units = 'GB' THEN resources.memory_deserved * 1000.0
                                  ELSE resources.memory_deserved
                                  END
                      ), 3)
           END AS total_memory_quota
FROM resources WHERE deleted_at IS NULL
GROUP BY org_unit_data_id
);

DROP VIEW IF EXISTS resources_deleted_view;
CREATE VIEW resources_deleted_view AS
(
SELECT org_unit_data_id,
       MAX(resources.deleted_at) AS resources_deleted_at
FROM resources WHERE deleted_at IS NOT NULL
GROUP BY org_unit_data_id
);

DROP VIEW IF EXISTS train_node_types_view;
CREATE VIEW train_node_types_view AS
(
SELECT json_agg(node_type.*) AS training_node_types_view, org_unit_train_node_type.org_unit_data_id,
       MAX(org_unit_train_node_type.created_at) AS train_node_type_created_at
FROM org_unit_train_node_type
         LEFT JOIN node_type ON node_type.id = org_unit_train_node_type.node_type_id
GROUP BY org_unit_train_node_type.org_unit_data_id
);

DROP VIEW IF EXISTS workspace_node_types_view;
CREATE VIEW workspace_node_types_view AS
(
SELECT json_agg(node_type.*) AS workspace_node_types_view, org_unit_workspace_node_type.org_unit_data_id,
    MAX(org_unit_workspace_node_type.created_at) AS workspace_node_type_created_at
FROM org_unit_workspace_node_type
         LEFT JOIN node_type ON node_type.id = org_unit_workspace_node_type.node_type_id
GROUP BY org_unit_workspace_node_type.org_unit_data_id
);

DROP VIEW IF EXISTS project_aggregated_status_view;
CREATE VIEW project_aggregated_status_view AS
(
SELECT project.parent_id,
           json_agg(json_build_object(
                   'id', project.id,
                   'name', project.name
                    )) as children_view,
           SUM(quota_status.gpu_allocated) AS gpu_allocated,
           SUM(quota_status.gpu_allocated_non_preemptible) AS gpu_allocated_non_preemptible,
           SUM(quota_status.gpu_requested) AS gpu_requested,
           SUM(quota_status.cpu_allocated) AS cpu_allocated,
           SUM(quota_status.cpu_allocated_non_preemptible) AS cpu_allocated_non_preemptible,
           SUM(quota_status.cpu_requested) AS cpu_requested,
           SUM(quota_status.memory_allocated) AS memory_allocated,
           SUM(quota_status.memory_allocated_non_preemptible) AS memory_allocated_non_preemptible,
           SUM(quota_status.memory_requested) AS memory_requested
    FROM quota_status
             LEFT JOIN project ON project.quota_status_id = quota_status.id
    WHERE project.deleted_at is null
    GROUP BY project.parent_id
);

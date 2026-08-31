-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS department_status_view;

CREATE VIEW department_status_view AS
(
WITH resources_agg AS (
    SELECT org_unit_data_id,
           json_agg(resources.*) AS resources_view,
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
    FROM resources
    GROUP BY org_unit_data_id
),
     train_node_types AS (
         SELECT json_agg(node_type.*) AS training_node_types_view , org_unit_train_node_type.org_unit_data_id
         FROM org_unit_train_node_type
                  LEFT JOIN node_type ON node_type.id = org_unit_train_node_type.node_type_id
         GROUP BY org_unit_train_node_type.org_unit_data_id
     ),
     workspace_node_types AS (
         SELECT json_agg(node_type.*) AS workspace_node_types_view , org_unit_workspace_node_type.org_unit_data_id
         FROM org_unit_workspace_node_type
                  LEFT JOIN node_type ON node_type.id = org_unit_workspace_node_type.node_type_id
         GROUP BY org_unit_workspace_node_type.org_unit_data_id
     ),
     projects_status_agg AS (
         SELECT project.parent_id,
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
         GROUP BY project.parent_id
     )
SELECT department.*,
       interactive_job_time_limit,
       interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration,
       training_job_time_limit,
       training_job_max_idle_duration,
       default_node_pool_list,
       resources_view,
       total_gpu_quota,
       total_cpu_quota,
       total_memory_quota,
       gpu_allocated,
       gpu_allocated_non_preemptible,
       gpu_requested,
       cpu_allocated,
       cpu_allocated_non_preemptible,
       cpu_requested,
       memory_allocated,
       memory_allocated_non_preemptible,
       memory_requested,
       training_node_types_view,
       workspace_node_types_view
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_agg ON resources_agg.org_unit_data_id = org_unit_data.id
         LEFT JOIN projects_status_agg ON projects_status_agg.parent_id = department.id
         LEFT JOIN train_node_types ON train_node_types.org_unit_data_id = org_unit_data.id
         LEFT JOIN workspace_node_types ON workspace_node_types.org_unit_data_id = org_unit_data.id
    );
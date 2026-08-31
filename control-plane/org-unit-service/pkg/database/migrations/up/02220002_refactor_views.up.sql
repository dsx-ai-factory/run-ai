-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS project_base_view;
DROP VIEW IF EXISTS department_base_view;
DROP VIEW IF EXISTS department_view;
DROP VIEW IF EXISTS project_view;
DROP VIEW IF EXISTS department_status_view;
DROP VIEW IF EXISTS project_sync_view;
DROP VIEW IF EXISTS department_sync_view;

CREATE VIEW department_view AS
(
SELECT department.*,
       interactive_job_time_limit,
       interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration,
       training_job_time_limit,
       training_job_max_idle_duration,
       default_node_pool_list,
       resources_json,
       total_gpu_quota,
       total_cpu_quota,
       total_memory_quota,
       training_node_types_view,
       workspace_node_types_view,
       overtime_data.id AS overtime_data_id,
       overtime_data.avg_gpu_allocation24h, overtime_data.avg_gpu_utilization24h,overtime_data.avg_gpu_memory_utilization24h,  overtime_data.avg24h_range_updated_at,
       overtime_data.avg_gpu_allocation7d, overtime_data.avg_gpu_utilization7d, overtime_data.avg_gpu_memory_utilization7d, overtime_data.avg7d_range_updated_at,
       overtime_data.avg_gpu_allocation30d, overtime_data.avg_gpu_utilization30d, overtime_data.avg_gpu_memory_utilization30d, overtime_data.avg30d_range_updated_at
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_aggregated_view ON resources_aggregated_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types_view ON train_node_types_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN workspace_node_types_view ON workspace_node_types_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN overtime_data ON overtime_data.org_unit_data_id = org_unit_data.id
    );

CREATE VIEW project_view AS
(
SELECT project.*,
       interactive_job_time_limit,
       interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration,
       training_job_time_limit,
       training_job_max_idle_duration,
       default_node_pool_list,
       interactive_job_time_limit_effective,
       interactive_job_max_idle_duration_effective,
       interactive_job_preempt_idle_duration_effective,
       training_job_time_limit_effective,
       training_job_max_idle_duration_effective,
       default_node_pool_list_effective,
       interactive_job_time_limit_effective_origin,
       interactive_job_max_idle_duration_effective_origin,
       interactive_job_preempt_idle_duration_effective_origin,
       training_job_time_limit_effective_origin,
       training_job_max_idle_duration_effective_origin,
       default_node_pool_list_effective_origin,
       training_node_types_effective_origin,
       workspace_node_types_effective_origin,
       resources_json,
       total_gpu_quota,
       total_cpu_quota,
       total_memory_quota,
       train_node_types_view.training_node_types_view,
       workspace_node_types_view.workspace_node_types_view,
       train_node_types_effective_view.training_node_types_view as training_node_types_view_effective,
       workspace_node_types_effective_view.workspace_node_types_view as workspace_node_types_view_effective,
       quota_status.gpu_allocated, quota_status.gpu_allocated_non_preemptible, quota_status.gpu_requested,
       quota_status.cpu_allocated, quota_status.cpu_allocated_non_preemptible, quota_status.cpu_requested,
       quota_status.memory_allocated, quota_status.memory_allocated_non_preemptible, quota_status.memory_requested,
       quota_status.node_pool_quota_statuses,
       overtime_data.id AS overtime_data_id,
       overtime_data.avg_gpu_allocation24h, overtime_data.avg_gpu_utilization24h, overtime_data.avg_gpu_memory_utilization24h, overtime_data.avg24h_range_updated_at,
       overtime_data.avg_gpu_allocation7d, overtime_data.avg_gpu_utilization7d, overtime_data.avg_gpu_memory_utilization7d, overtime_data.avg7d_range_updated_at,
       overtime_data.avg_gpu_allocation30d, overtime_data.avg_gpu_utilization30d, overtime_data.avg_gpu_memory_utilization30d, overtime_data.avg30d_range_updated_at
FROM project
         INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_aggregated_view ON resources_aggregated_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types_view ON train_node_types_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN workspace_node_types_view ON workspace_node_types_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types_view as train_node_types_effective_view ON train_node_types_effective_view.org_unit_data_id = COALESCE(org_unit_data.training_node_types_effective_origin, org_unit_data.id)
         LEFT JOIN workspace_node_types_view as workspace_node_types_effective_view ON workspace_node_types_effective_view.org_unit_data_id = COALESCE(org_unit_data.workspace_node_types_effective_origin, org_unit_data.id)
         LEFT JOIN quota_status ON project.quota_status_id = quota_status.id
         LEFT JOIN overtime_data ON overtime_data.org_unit_data_id = org_unit_data.id
    );

CREATE VIEW department_status_view AS
(
SELECT department.*,
       interactive_job_time_limit,
       interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration,
       training_job_time_limit,
       training_job_max_idle_duration,
       default_node_pool_list,
       resources_json,
       total_gpu_quota,
       total_cpu_quota,
       total_memory_quota,
       children_view,
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
       workspace_node_types_view,
       overtime_data.id AS overtime_data_id,
       overtime_data.avg_gpu_allocation24h, overtime_data.avg_gpu_utilization24h, overtime_data.avg_gpu_memory_utilization24h, overtime_data.avg24h_range_updated_at,
       overtime_data.avg_gpu_allocation7d, overtime_data.avg_gpu_utilization7d, overtime_data.avg_gpu_memory_utilization7d, overtime_data.avg7d_range_updated_at,
       overtime_data.avg_gpu_allocation30d, overtime_data.avg_gpu_utilization30d, overtime_data.avg_gpu_memory_utilization30d, overtime_data.avg30d_range_updated_at
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_aggregated_view ON resources_aggregated_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN project_aggregated_status_view ON project_aggregated_status_view.parent_id = department.id
         LEFT JOIN train_node_types_view ON train_node_types_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN workspace_node_types_view ON workspace_node_types_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN overtime_data ON overtime_data.org_unit_data_id = org_unit_data.id
    );

CREATE VIEW project_sync_view AS
(
SELECT project.*,
       interactive_job_time_limit_effective as interactive_job_time_limit,
       interactive_job_max_idle_duration_effective as interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration_effective as interactive_job_preempt_idle_duration,
       training_job_time_limit_effective as training_job_time_limit,
       training_job_max_idle_duration_effective as training_job_max_idle_duration,
       default_node_pool_list_effective as default_node_pool_list,
       resources_json,
       training_node_types_view,
       workspace_node_types_view,
       GREATEST(
               COALESCE(project.updated_at, '1970-01-01')
           , COALESCE(org_unit_data.updated_at, '1970-01-01')
           , COALESCE(resources_aggregated_view.resources_updated_at, '1970-01-01')
           , COALESCE(resources_deleted_view.resources_deleted_at, '1970-01-01')
           , COALESCE(train_node_types_effective_view.train_node_type_created_at, '1970-01-01')
           , COALESCE(workspace_node_types_effective_view.workspace_node_type_created_at, '1970-01-01')
       ) AS calculated_updated_at
FROM project
         INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_aggregated_view ON resources_aggregated_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_deleted_view ON resources_deleted_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types_view as train_node_types_effective_view ON train_node_types_effective_view.org_unit_data_id = COALESCE(org_unit_data.training_node_types_effective_origin, org_unit_data.id)
         LEFT JOIN workspace_node_types_view as workspace_node_types_effective_view ON workspace_node_types_effective_view.org_unit_data_id = COALESCE(org_unit_data.workspace_node_types_effective_origin, org_unit_data.id)
    );

CREATE VIEW department_sync_view AS
(
SELECT department.*,
       resources_json,
       GREATEST(
               COALESCE(department.updated_at, '1970-01-01')
           , COALESCE(org_unit_data.updated_at, '1970-01-01')
           , COALESCE(resources_deleted_view.resources_deleted_at, '1970-01-01')
           , COALESCE(resources_aggregated_view.resources_updated_at, '1970-01-01')
       ) AS calculated_updated_at
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_aggregated_view ON resources_aggregated_view.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_deleted_view ON resources_deleted_view.org_unit_data_id = org_unit_data.id
    );
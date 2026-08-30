-- Drop functions and types (re-runnable)
DROP FUNCTION IF EXISTS get_sync_projects(TEXT, UUID);
DROP FUNCTION IF EXISTS get_sync_departments(TEXT, UUID);
DROP TYPE IF EXISTS project_sync_type CASCADE;
DROP TYPE IF EXISTS department_sync_type CASCADE;

-- Drop any pre-existing views so CREATE VIEW below is re-runnable after partial failure
DROP VIEW IF EXISTS project_sync_view CASCADE;
DROP VIEW IF EXISTS department_sync_view CASCADE;

-- Recreate original views (from 02240001)
CREATE VIEW project_sync_view AS
(
WITH resources_agg AS (
    SELECT org_unit_data_id,
           MAX(resources.updated_at) AS resources_updated_at,
           json_agg(resources.* ORDER BY id) AS resources_json,
           CASE
               WHEN COUNT(*) FILTER (WHERE resources.gpu_deserved IS NULL) > 0 THEN NULL
               WHEN COUNT(*) FILTER (WHERE resources.gpu_deserved = -1) > 0 THEN -1
               ELSE SUM(resources.gpu_deserved)
               END AS total_gpu_quota,
           CASE
               WHEN COUNT(*) FILTER (WHERE resources.cpu_deserved IS NULL) > 0 THEN NULL
               WHEN COUNT(*) FILTER (WHERE resources.cpu_deserved = -1) > 0 THEN -1
               ELSE SUM(resources.cpu_deserved)
               END AS total_cpu_quota,
           CASE
               WHEN COUNT(*) FILTER (WHERE resources.memory_deserved IS NULL) > 0 THEN NULL
               WHEN COUNT(*) FILTER (WHERE resources.memory_deserved = -1) > 0 THEN -1
               ELSE ROUND(SUM(
                                  resources.memory_deserved *
                                  CASE
                                      WHEN resources.memory_units IS NULL OR resources.memory_units = 'Mib' THEN 1.048576
                                      WHEN resources.memory_units = 'GB' OR resources.memory_units = 'Gb' THEN 1000.0
                                      ELSE 1.0
                                      END
                          ), 3)
               END AS total_memory_quota
    FROM resources WHERE deleted_at IS NULL
    GROUP BY org_unit_data_id
),
     resources_deleted AS (
         SELECT org_unit_data_id,
                MAX(resources.deleted_at) AS resources_deleted_at
         FROM resources WHERE deleted_at IS NOT NULL
         GROUP BY org_unit_data_id
     ),
     train_node_types_effective AS (
         SELECT json_agg(node_type.*) AS training_node_types_view_effective, org_unit_train_node_type.org_unit_data_id,
                MAX(org_unit_train_node_type.created_at) AS train_node_type_created_at
         FROM org_unit_train_node_type
                  LEFT JOIN node_type ON node_type.id = org_unit_train_node_type.node_type_id
         GROUP BY org_unit_train_node_type.org_unit_data_id
     ),
     workspace_node_types_effective AS (
         SELECT json_agg(node_type.*) AS workspace_node_types_view_effective, org_unit_workspace_node_type.org_unit_data_id,
                MAX(org_unit_workspace_node_type.created_at) AS workspace_node_type_created_at
         FROM org_unit_workspace_node_type
                  LEFT JOIN node_type ON node_type.id = org_unit_workspace_node_type.node_type_id
         GROUP BY org_unit_workspace_node_type.org_unit_data_id
     )
SELECT project.*,
       interactive_job_time_limit_effective as interactive_job_time_limit,
       interactive_job_max_idle_duration_effective as interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration_effective as interactive_job_preempt_idle_duration,
       training_job_time_limit_effective as training_job_time_limit,
       training_job_max_idle_duration_effective as training_job_max_idle_duration,
       default_node_pool_list_effective as default_node_pool_list,
       resources_json,
       training_node_types_view_effective as training_node_types_view,
       workspace_node_types_view_effective as workspace_node_types_view,
       GREATEST(
               COALESCE(project.updated_at, '1970-01-01')
           , COALESCE(org_unit_data.updated_at, '1970-01-01')
           , COALESCE(resources_agg.resources_updated_at, '1970-01-01')
           , COALESCE(resources_deleted.resources_deleted_at, '1970-01-01')
           , COALESCE(train_node_types_effective.train_node_type_created_at, '1970-01-01')
           , COALESCE(workspace_node_types_effective.workspace_node_type_created_at, '1970-01-01')
       ) AS calculated_updated_at
FROM project
         INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_agg ON resources_agg.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_deleted ON resources_deleted.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types_effective ON train_node_types_effective.org_unit_data_id = COALESCE(org_unit_data.training_node_types_effective_origin, org_unit_data.id)
         LEFT JOIN workspace_node_types_effective ON workspace_node_types_effective.org_unit_data_id = COALESCE(org_unit_data.workspace_node_types_effective_origin, org_unit_data.id)
    );

CREATE VIEW department_sync_view AS
(
WITH resources_agg AS (
    SELECT org_unit_data_id,
           MAX(resources.updated_at) AS resources_updated_at,
           json_agg(resources.* ORDER BY id) AS resources_json,
           CASE
               WHEN COUNT(*) FILTER (WHERE resources.gpu_deserved IS NULL) > 0 THEN NULL
               WHEN COUNT(*) FILTER (WHERE resources.gpu_deserved = -1) > 0 THEN -1
               ELSE SUM(resources.gpu_deserved)
               END AS total_gpu_quota,
           CASE
               WHEN COUNT(*) FILTER (WHERE resources.cpu_deserved IS NULL) > 0 THEN NULL
               WHEN COUNT(*) FILTER (WHERE resources.cpu_deserved = -1) > 0 THEN -1
               ELSE SUM(resources.cpu_deserved)
               END AS total_cpu_quota,
           CASE
               WHEN COUNT(*) FILTER (WHERE resources.memory_deserved IS NULL) > 0 THEN NULL
               WHEN COUNT(*) FILTER (WHERE resources.memory_deserved = -1) > 0 THEN -1
               ELSE ROUND(SUM(
                                  resources.memory_deserved *
                                  CASE
                                      WHEN resources.memory_units IS NULL OR resources.memory_units = 'Mib' THEN 1.048576
                                      WHEN resources.memory_units = 'GB' OR resources.memory_units = 'Gb' THEN 1000.0
                                      ELSE 1.0
                                      END
                          ), 3)
               END AS total_memory_quota
    FROM resources WHERE deleted_at IS NULL
    GROUP BY org_unit_data_id
),
     resources_deleted AS (
         SELECT org_unit_data_id,
                MAX(resources.deleted_at) AS resources_deleted_at
         FROM resources WHERE deleted_at IS NOT NULL
         GROUP BY org_unit_data_id
     )
SELECT department.*,
       resources_json,
       GREATEST(
               COALESCE(department.updated_at, '1970-01-01')
           , COALESCE(org_unit_data.updated_at, '1970-01-01')
           , COALESCE(resources_deleted.resources_deleted_at, '1970-01-01')
           , COALESCE(resources_agg.resources_updated_at, '1970-01-01')
       ) AS calculated_updated_at
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_agg ON resources_agg.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_deleted ON resources_deleted.org_unit_data_id = org_unit_data.id
    );

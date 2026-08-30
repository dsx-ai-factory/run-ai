ALTER TABLE IF EXISTS org_unit_data ADD COLUMN IF NOT EXISTS sync_failed BOOLEAN NOT NULL DEFAULT FALSE;

DROP VIEW IF EXISTS project_sync_view;
DROP VIEW IF EXISTS department_sync_view;

CREATE VIEW project_sync_view AS
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
     train_node_types_effective AS (
         SELECT json_agg(node_type.*) AS training_node_types_view_effective, org_unit_train_node_type.org_unit_data_id
         FROM org_unit_train_node_type
                  LEFT JOIN node_type ON node_type.id = org_unit_train_node_type.node_type_id
         GROUP BY org_unit_train_node_type.org_unit_data_id
     ),
     workspace_node_types_effective AS (
         SELECT json_agg(node_type.*) AS workspace_node_types_view_effective, org_unit_workspace_node_type.org_unit_data_id
         FROM org_unit_workspace_node_type
                  LEFT JOIN node_type ON node_type.id = org_unit_workspace_node_type.node_type_id
         GROUP BY org_unit_workspace_node_type.org_unit_data_id
     )
SELECT project.*,
       sync_failed,
       interactive_job_time_limit_effective as interactive_job_time_limit,
       interactive_job_max_idle_duration_effective as interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration_effective as interactive_job_preempt_idle_duration,
       training_job_time_limit_effective as training_job_time_limit,
       training_job_max_idle_duration_effective as training_job_max_idle_duration,
       default_node_pool_list_effective as default_node_pool_list,
       resources_view,
       training_node_types_view_effective as training_node_types_view,
       workspace_node_types_view_effective as workspace_node_types_view
FROM project
         INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_agg ON resources_agg.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types_effective ON train_node_types_effective.org_unit_data_id = COALESCE(org_unit_data.training_node_types_effective_origin, org_unit_data.id)
         LEFT JOIN workspace_node_types_effective ON workspace_node_types_effective.org_unit_data_id = COALESCE(org_unit_data.workspace_node_types_effective_origin, org_unit_data.id)
    );

CREATE VIEW department_sync_view AS
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
     )
SELECT department.*,
       sync_failed,
       interactive_job_time_limit,
       interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration,
       training_job_time_limit,
       training_job_max_idle_duration,
       default_node_pool_list,
       resources_view,
       training_node_types_view,
       workspace_node_types_view
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_agg ON resources_agg.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types ON train_node_types.org_unit_data_id = org_unit_data.id
         LEFT JOIN workspace_node_types ON workspace_node_types.org_unit_data_id = org_unit_data.id
    );

DROP VIEW IF EXISTS project_view;
DROP VIEW IF EXISTS department_view;
DROP VIEW IF EXISTS project_base_view;
DROP VIEW IF EXISTS department_base_view;

CREATE VIEW project_view AS
(
SELECT p.*,
       json_build_object(
               'gpu_allocated', qs.gpu_allocated,
               'gpu_allocated_non_preemptible', qs.gpu_allocated_non_preemptible,
               'gpu_requested', qs.gpu_requested,
               'cpu_allocated', qs.cpu_allocated,
               'cpu_allocated_non_preemptible', qs.cpu_allocated_non_preemptible,
               'cpu_requested', qs.cpu_requested,
               'memory_allocated', qs.memory_allocated,
               'memory_allocated_non_preemptible', qs.memory_allocated_non_preemptible,
               'memory_requested', qs.memory_requested,
               'node_pool_quota_statuses', qs.node_pool_quota_statuses
           ) AS quota_status_raw,
       json_build_object(
               'id', oud.id,
               'created_at', oud.created_at,
               'updated_at', oud.updated_at,
               'deleted_at', oud.deleted_at,
               'created_by', oud.created_by,
               'updated_by', oud.updated_by,
               'deleted_by', oud.deleted_by,
               'interactive_job_time_limit', oud.interactive_job_time_limit,
               'interactive_job_max_idle_duration', oud.interactive_job_max_idle_duration,
               'interactive_job_preempt_idle_duration', oud.interactive_job_preempt_idle_duration,
               'training_job_time_limit', oud.training_job_time_limit,
               'training_job_max_idle_duration', oud.training_job_max_idle_duration,
               'default_node_pool_list', oud.default_node_pool_list,
               'node_type_defined', oud.node_type_defined,
               'resources', json_agg(r.*),
               'training_node_types', json_agg(json_build_object('id', tnt.node_type_id, 'name', nt.name, 'cluster_id', nt.cluster_id, 'tenant_id', nt.tenant_id)),
               'workspace_node_types', json_agg(json_build_object('id', wnt.node_type_id, 'name', nt.name, 'cluster_id', nt.cluster_id, 'tenant_id', nt.tenant_id))
           ) AS org_unit_data_raw
FROM project p
         INNER JOIN org_unit_data oud ON p.org_unit_data_id = oud.id
         LEFT JOIN resources r ON r.org_unit_data_id = oud.id
         LEFT JOIN org_unit_train_node_type tnt ON tnt.org_unit_data_id = oud.id
         LEFT JOIN org_unit_workspace_node_type wnt ON wnt.org_unit_data_id = oud.id
         LEFT JOIN node_type nt ON nt.id = tnt.node_type_id OR nt.id = wnt.node_type_id
         LEFT JOIN quota_status qs ON p.quota_status_id = qs.id
GROUP BY p.id, oud.id, qs.id
    );

CREATE VIEW department_view AS
(
SELECT d.*,
       json_build_object(
               'id', oud.id,
               'created_at', oud.created_at,
               'updated_at', oud.updated_at,
               'deleted_at', oud.deleted_at,
               'created_by', oud.created_by,
               'updated_by', oud.updated_by,
               'deleted_by', oud.deleted_by,
               'interactive_job_time_limit', oud.interactive_job_time_limit,
               'interactive_job_max_idle_duration', oud.interactive_job_max_idle_duration,
               'interactive_job_preempt_idle_duration', oud.interactive_job_preempt_idle_duration,
               'training_job_time_limit', oud.training_job_time_limit,
               'training_job_max_idle_duration', oud.training_job_max_idle_duration,
               'default_node_pool_list', oud.default_node_pool_list,
               'node_type_defined', oud.node_type_defined,
               'resources', json_agg(r.*),
               'training_node_types', json_agg(json_build_object('id', tnt.node_type_id, 'name', nt.name, 'cluster_id', nt.cluster_id, 'tenant_id', nt.tenant_id)),
               'workspace_node_types', json_agg(json_build_object('id', wnt.node_type_id, 'name', nt.name, 'cluster_id', nt.cluster_id, 'tenant_id', nt.tenant_id))
           ) AS org_unit_data_raw
FROM department d
         INNER JOIN org_unit_data oud ON d.org_unit_data_id = oud.id
         LEFT JOIN resources r ON r.org_unit_data_id = oud.id
         LEFT JOIN org_unit_train_node_type tnt ON tnt.org_unit_data_id = oud.id
         LEFT JOIN org_unit_workspace_node_type wnt ON wnt.org_unit_data_id = oud.id
         LEFT JOIN node_type nt ON nt.id = tnt.node_type_id OR nt.id = wnt.node_type_id
GROUP BY d.id, oud.id
    );

CREATE VIEW department_base_view AS
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
       training_node_types_view,
       workspace_node_types_view
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_agg ON resources_agg.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types ON train_node_types.org_unit_data_id = org_unit_data.id
         LEFT JOIN workspace_node_types ON workspace_node_types.org_unit_data_id = org_unit_data.id
    );

CREATE VIEW project_base_view AS
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
SELECT project.*,
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
       training_node_types_view,
       workspace_node_types_view,
       quota_status.gpu_allocated, quota_status.gpu_allocated_non_preemptible, quota_status.gpu_requested,
       quota_status.cpu_allocated, quota_status.cpu_allocated_non_preemptible, quota_status.cpu_requested,
       quota_status.memory_allocated, quota_status.memory_allocated_non_preemptible, quota_status.memory_requested,
       quota_status.node_pool_quota_statuses
FROM project
         INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources_agg ON resources_agg.org_unit_data_id = org_unit_data.id
         LEFT JOIN train_node_types ON train_node_types.org_unit_data_id = org_unit_data.id
         LEFT JOIN workspace_node_types ON workspace_node_types.org_unit_data_id = org_unit_data.id
         LEFT JOIN quota_status ON project.quota_status_id = quota_status.id
    );

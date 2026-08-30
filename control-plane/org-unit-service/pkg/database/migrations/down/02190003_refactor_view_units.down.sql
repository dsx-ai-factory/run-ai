DROP VIEW IF EXISTS project_view;
DROP VIEW IF EXISTS department_view;

CREATE VIEW department_view AS
(
SELECT department.*,
       interactive_job_time_limit,
       interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration,
       training_job_time_limit,
       training_job_max_idle_duration,
       default_node_pool_list,
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
           ELSE SUM(resources.memory_deserved)
           END AS total_memory_quota,
       json_agg(train_node_type.*) AS training_node_types_view,
       json_agg(workspace_node_type.*) AS workspace_node_types_view
FROM department
         INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources ON resources.org_unit_data_id = org_unit_data.id
         LEFT JOIN (SELECT node_type.*,org_unit_train_node_type.org_unit_data_id FROM node_type
                                                                                          LEFT JOIN org_unit_train_node_type ON node_type.id = org_unit_train_node_type.node_type_id) AS train_node_type ON train_node_type.org_unit_data_id = org_unit_data.id
         LEFT JOIN (SELECT node_type.*,org_unit_workspace_node_type.org_unit_data_id FROM node_type
                                                                                              LEFT JOIN org_unit_workspace_node_type ON node_type.id = org_unit_workspace_node_type.node_type_id) AS workspace_node_type ON workspace_node_type.org_unit_data_id = org_unit_data.id
GROUP BY department.id, org_unit_data.id
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
           ELSE SUM(resources.memory_deserved)
           END AS total_memory_quota,
       json_agg(train_node_type.*) AS training_node_types_view,
       json_agg(workspace_node_type.*) AS workspace_node_types_view,
       quota_status.gpu_allocated, quota_status.gpu_allocated_non_preemptible, quota_status.gpu_requested,
       quota_status.cpu_allocated, quota_status.cpu_allocated_non_preemptible, quota_status.cpu_requested,
       quota_status.memory_allocated, quota_status.memory_allocated_non_preemptible, quota_status.memory_requested,
       quota_status.node_pool_quota_statuses
FROM project
         INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
         LEFT JOIN resources ON resources.org_unit_data_id = org_unit_data.id
         LEFT JOIN (SELECT node_type.*,org_unit_train_node_type.org_unit_data_id FROM node_type
                                                                                          LEFT JOIN org_unit_train_node_type ON node_type.id = org_unit_train_node_type.node_type_id) AS train_node_type ON train_node_type.org_unit_data_id = org_unit_data.id
         LEFT JOIN (SELECT node_type.*,org_unit_workspace_node_type.org_unit_data_id FROM node_type
                                                                                              LEFT JOIN org_unit_workspace_node_type ON node_type.id = org_unit_workspace_node_type.node_type_id) AS workspace_node_type ON workspace_node_type.org_unit_data_id = org_unit_data.id
         LEFT JOIN quota_status ON project.quota_status_id = quota_status.id
GROUP BY project.id, org_unit_data.id, quota_status.id
    );
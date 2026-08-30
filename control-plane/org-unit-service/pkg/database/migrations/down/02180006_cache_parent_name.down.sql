ALTER TABLE project DROP COLUMN IF EXISTS parent_name;

DROP VIEW IF EXISTS project_view;
CREATE VIEW project_view AS
(
SELECT p.*,
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
GROUP BY p.id, oud.id
    );
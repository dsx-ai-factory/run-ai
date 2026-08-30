DROP VIEW IF EXISTS project_node_pool_resources_view;
CREATE VIEW project_node_pool_resources_view AS
(
SELECT resources.id,
       resources.created_at,
       resources.updated_at,
       resources.created_by,
       resources.updated_by,
       resources.node_pool_id,
       resources.node_pool_name,
       resources.gpu_deserved,
       resources.gpu_limit,
       resources.gpu_over_quota_weight,
       resources.cpu_deserved,
       resources.cpu_limit,
       resources.cpu_over_quota_weight,
       resources.memory_deserved,
       resources.memory_limit,
       resources.memory_over_quota_weight,
       resources.memory_units,
       COALESCE(resources.priority, 'MediumLow') as rank,
       project.id as org_unit_id,
       project.name as org_unit_name,
       project.parent_id,
       project.cluster_id,
       project.tenant_id
FROM resources
    INNER JOIN project ON resources.org_unit_data_id = project.org_unit_data_id
    WHERE project.deleted_at is null
);


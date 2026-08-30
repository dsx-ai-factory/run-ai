WITH mapping AS (
    SELECT name, "group", version
    FROM (
             VALUES
                 ('SeldonDeployment', 'machinelearning.seldon.io', 'v1alpha2'),
                 ('MPIJob', 'kubeflow.org', 'v2beta1'),
                 ('RayCluster', 'ray.io', 'v1alpha1'),
                 ('RayService', 'ray.io', 'v1alpha1'),
                 ('RayJob', 'ray.io', 'v1alpha1')
         ) AS u(name, "group", version)
)
INSERT INTO workload_type_resource_interfaces (id, tenant_id, scope_type, workload_type_id, version, resource_interface, created_by)
SELECT gen_random_uuid(),
       NULL,
       'system',
       workload_types.id,
       m.version,
       jsonb_build_object(
               'spec', jsonb_build_object(
               'structureDefinition', jsonb_build_object(
                       'rootComponent', jsonb_build_object(
                               'kind', jsonb_build_object(
                                       'group', m."group",
                                       'kind', m.name,
                                       'version', m.version
                                       )
                               )
                       )
               )
       ),
       'Run.ai'
FROM mapping m, workload_types
WHERE workload_types.name = m.name
  AND NOT EXISTS (
    SELECT 1 FROM workload_type_resource_interfaces
    WHERE workload_type_id = workload_types.id
      AND version = m.version
);

DROP INDEX IF EXISTS workload_type_status_cluster;
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_status_cluster_version ON workload_type_statuses(workload_type_id,version,cluster_id,tenant_id);
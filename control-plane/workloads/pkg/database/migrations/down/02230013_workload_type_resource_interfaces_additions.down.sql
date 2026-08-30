DROP INDEX IF EXISTS workload_type_status_cluster_version;
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_status_cluster ON workload_type_statuses(workload_type_id,cluster_id,tenant_id);

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
DELETE FROM workload_type_resource_interfaces r
USING mapping m, workload_types
WHERE workload_types.id = r.workload_type_id
  AND workload_types.name = m.name
  AND workload_types."group" = m."group"
  AND r.version = m.version;

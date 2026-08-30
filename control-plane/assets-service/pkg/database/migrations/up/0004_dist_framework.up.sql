ALTER TABLE workloads ADD COLUMN if not exists dist_framework text;
DROP INDEX workloads_by_reference;
CREATE UNIQUE INDEX IF NOT EXISTS workloads_by_reference ON workloads (tenant_id,cluster_id,project_id,workload_type,name, dist_framework);

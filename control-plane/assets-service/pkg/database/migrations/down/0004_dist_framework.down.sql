ALTER TABLE workloads DROP COLUMN if exists dist_framework;
DROP INDEX if exists workloads_by_reference;
CREATE UNIQUE INDEX IF NOT EXISTS workloads_by_reference ON workloads (tenant_id,cluster_id,project_id,workload_type,name);

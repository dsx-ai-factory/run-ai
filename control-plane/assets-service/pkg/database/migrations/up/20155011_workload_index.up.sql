---- Redo of 0011_workload_index.up.sql in idempotent in case it didn't run
CREATE INDEX IF NOT EXISTS workloads_by_name ON workloads (tenant_id, name);
CREATE INDEX IF NOT EXISTS workloads_by_type ON workloads (tenant_id, workload_type);
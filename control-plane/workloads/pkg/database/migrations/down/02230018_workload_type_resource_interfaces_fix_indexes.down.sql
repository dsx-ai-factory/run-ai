-- Drop the partial unique indexes introduced in the up migration
DROP INDEX IF EXISTS workload_type_resource_interfaces_version_with_tenant;
DROP INDEX IF EXISTS workload_type_resource_interfaces_version_null_tenant;

-- Restore the original unique index that includes tenant_id
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_resource_interfaces_version ON workload_type_resource_interfaces(workload_type_id,version,scope_type,tenant_id);


-- Drop the old unique index that allowed multiple NULL tenant_id rows
DROP INDEX IF EXISTS workload_type_resource_interfaces_version;

-- Enforce uniqueness when tenant_id IS NULL (treat NULL as a real value for uniqueness)
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_resource_interfaces_version_null_tenant
    ON workload_type_resource_interfaces (workload_type_id, version, scope_type)
    WHERE tenant_id IS NULL;

-- Enforce uniqueness per-tenant when tenant_id IS NOT NULL
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_resource_interfaces_version_with_tenant
    ON workload_type_resource_interfaces (workload_type_id, version, scope_type, tenant_id)
    WHERE tenant_id IS NOT NULL;

DROP TRIGGER IF EXISTS enforce_tenant_name_alias_cross_uniqueness ON tenants;
DROP FUNCTION IF EXISTS tenant_name_alias_cross_uniqueness();
DROP INDEX IF EXISTS idx_tenants_alias;
ALTER TABLE tenants DROP COLUMN IF EXISTS alias;
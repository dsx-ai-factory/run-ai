ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS tenant_id text;
CREATE INDEX IF NOT EXISTS roles_tenant_idx ON roles(tenant_id);


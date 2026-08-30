DROP INDEX IF EXISTS assets_metadata_by_kind_name_tenant_private;
DROP INDEX IF EXISTS assets_metadata_by_kind_name_tenant;

CREATE UNIQUE INDEX assets_metadata_by_kind_name_tenant
    ON assets_metadata (kind, tenant_id, name)
    WHERE scope = 'tenant'::text;

ALTER TABLE assets_metadata DROP COLUMN IF EXISTS private;
ALTER TABLE reconciler_versions DROP CONSTRAINT if exists fk_tenant_id;

ALTER TABLE idps DROP CONSTRAINT if exists fk_tenant_id;

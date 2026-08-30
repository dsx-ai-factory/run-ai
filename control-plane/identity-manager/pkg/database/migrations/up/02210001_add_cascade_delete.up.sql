ALTER TABLE reconciler_versions DROP CONSTRAINT if exists fk_tenant_id;
ALTER TABLE reconciler_versions ADD CONSTRAINT fk_tenant_id  FOREIGN KEY (tenant_id) REFERENCES realms(tenant_id) ON DELETE CASCADE;

ALTER TABLE idps DROP CONSTRAINT if exists fk_tenant_id;
ALTER TABLE idps ADD CONSTRAINT fk_tenant_id  FOREIGN KEY (tenant_id) REFERENCES realms(tenant_id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS settings (
    id                          UUID NOT NULL PRIMARY KEY,
    tenant_id                   TEXT NOT NULL,
    key                         TEXT NOT NULL,
    value                       JSONB,
    updated_at                  TIMESTAMPTZ,
    updated_by_customer         BOOLEAN NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES realms(tenant_id)
);
ALTER TABLE settings DROP CONSTRAINT if exists unique_setting;
ALTER TABLE settings DROP CONSTRAINT if exists fk_tenant_id_realm;
ALTER TABLE settings ADD CONSTRAINT unique_setting UNIQUE (tenant_id, key);
ALTER TABLE settings ADD CONSTRAINT fk_tenant_id_realm  FOREIGN KEY (tenant_id) REFERENCES realms(tenant_id) ON DELETE CASCADE;

INSERT INTO identity_manager.settings (id, tenant_id, key, value, updated_at, updated_by_customer)
SELECT 
    id,
    tenant_id,
    setting_key AS key,
    setting_value AS value,
    updated_at,
    updated_by_customer
FROM identity_manager.realm_settings
WHERE NOT EXISTS (SELECT 1 FROM identity_manager.settings LIMIT 1)
ON CONFLICT (tenant_id, key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at,
    updated_by_customer = EXCLUDED.updated_by_customer;

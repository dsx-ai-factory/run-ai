CREATE TABLE IF NOT EXISTS tenant_settings (
    tenant_id   INTEGER,
    key         TEXT,
    value       TEXT,
    type        TEXT,

    CONSTRAINT tenant_settings_tenant_id_key_pk PRIMARY KEY(tenant_id, key)
);

CREATE INDEX IF NOT EXISTS tenant_settings_tenant_id_idx ON tenant_settings (tenant_id);
CREATE INDEX IF NOT EXISTS tenant_settings_key_idx ON tenant_settings (key);

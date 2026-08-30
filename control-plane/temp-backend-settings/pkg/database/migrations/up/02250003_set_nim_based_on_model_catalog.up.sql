INSERT INTO tenant_settings (tenant_id, key, value, type)
SELECT tenant_id, 'tenant.enable_nim', 'false', 'Boolean'
FROM tenant_settings
WHERE key = 'tenant.enable_model_catalog'
  AND value = 'false'
  AND tenant_id NOT IN (
    SELECT tenant_id
    FROM tenant_settings
    WHERE key = 'tenant.enable_nim'
  );

UPDATE realms as r
SET exclude_groups_from_tokens = true
FROM realm_settings as rs
WHERE r.tenant_id = rs.tenant_id AND rs.setting_key = 'excludeGroupsFromToken' AND rs.setting_value->>'enabled' = 'true';
UPDATE tenants SET mimir_id = id WHERE mimir_id IS NULL AND smg = true;

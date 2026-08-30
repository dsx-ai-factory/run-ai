ALTER TABLE IF EXISTS permissions DROP CONSTRAINT IF EXISTS fk_permission_set_permissions;
ALTER TABLE IF EXISTS permissions DROP COLUMN IF EXISTS permission_set_id;

DROP INDEX IF EXISTS idx_role_permissions_resource_type;
DROP TABLE IF EXISTS role_permissions;
ALTER TABLE IF EXISTS permission_sets DROP CONSTRAINT IF EXISTS permission_sets_name_key;
ALTER TABLE IF EXISTS permission_sets DROP COLUMN IF EXISTS created_by;
ALTER TABLE IF EXISTS permission_sets DROP COLUMN IF EXISTS updated_by;

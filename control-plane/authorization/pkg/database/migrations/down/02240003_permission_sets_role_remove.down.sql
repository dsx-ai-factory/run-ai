ALTER TABLE IF EXISTS permissions ADD COLUMN IF NOT EXISTS permission_set_id UUID;
ALTER TABLE IF EXISTS permissions ADD CONSTRAINT fk_permission_set_permissions FOREIGN KEY(permission_set_id) REFERENCES permission_sets(id);

ALTER TABLE IF EXISTS disabled_roles ADD CONSTRAINT unique_scope_disabled_role UNIQUE (scope_id, scope_type, role_id);

DROP INDEX IF EXISTS pool_to_org_unit_data;

UPDATE resources SET deleted_at = null, deleted_by = '' where deleted_by='runai_duplication_migration'

CREATE UNIQUE INDEX IF NOT EXISTS pool_to_org_unit_data ON resources(deleted_at,org_unit_data_id,node_pool_id);

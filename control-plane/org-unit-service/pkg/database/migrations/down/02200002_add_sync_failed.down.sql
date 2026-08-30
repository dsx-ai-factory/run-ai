ALTER TABLE IF EXISTS org_unit_data DROP COLUMN IF EXISTS sync_failed;

DROP VIEW IF EXISTS project_sync_view;
DROP VIEW IF EXISTS department_sync_view;

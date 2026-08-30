---- Redo of 0009_workload_template.down.sql in idempotent in case it didn't run
DROP TRIGGER IF EXISTS validate_asset_name ON assets_metadata;

update assets_metadata set kind = 'workspace-template' where kind = 'workload-template';
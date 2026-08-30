DROP VIEW IF EXISTS external_workloads_type_config_view;

ALTER TABLE workload_types RENAME COLUMN type TO source;
ALTER TABLE workload_types DROP COLUMN IF EXISTS updated_by;
ALTER TABLE workload_types DROP COLUMN IF EXISTS updated_at;

UPDATE workload_types
SET created_by = 'Run.ai'
WHERE created_by IS NULL OR created_by = '';
UPDATE external_workloads_type_config
SET created_by = 'Run.ai'
WHERE created_by IS NULL OR created_by = '';
UPDATE external_workloads_type_config
SET updated_by = 'Run.ai'
WHERE updated_by IS NULL OR updated_by = '';

CREATE VIEW external_workloads_type_config_view AS
SELECT
    workload_types.id,
    workload_types.name,
    workload_types.source,
    config.created_at,
    config.created_by,
    config.updated_at,
    config.updated_by,
    COALESCE(category.id, system_category.id) AS category_id,
    COALESCE(priority.id, system_priority.id) AS priority_id,
    COALESCE(category.name, system_category.name) AS category_name,
    COALESCE(priority.name, system_priority.name) AS priority_name,
    config.scope_type,
    config.tenant_id
FROM workload_types
         INNER JOIN external_workloads_type_config config
                   ON config.workload_type_id = workload_types.id
         LEFT JOIN external_workloads_type_config config_system
                   ON config_system.workload_type_id = workload_types.id AND config_system.scope_type = 'system'
         LEFT JOIN workloads_categories category
                   ON category.id = config.category_id
         LEFT JOIN priorities priority
                   ON priority.id = config.priority_id
         LEFT JOIN workloads_categories system_category
                   ON system_category.id = config_system.category_id
         LEFT JOIN priorities system_priority
                   ON system_priority.id = config_system.priority_id;
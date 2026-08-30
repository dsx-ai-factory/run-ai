-- Drop the dependent view first so the table/column renames don't fail on dependency.
DROP VIEW IF EXISTS external_workloads_type_config_view;

-- Rename table: workload_type_resource_interfaces -> workload_type_kartas
ALTER TABLE IF EXISTS workload_type_resource_interfaces RENAME TO workload_type_kartas;

-- Rename column: resource_interface -> karta. RENAME COLUMN has no IF EXISTS form,
-- so wrap in DO/EXCEPTION to be re-runnable.
DO $$
BEGIN
    ALTER TABLE workload_type_kartas RENAME COLUMN resource_interface TO karta;
EXCEPTION
    WHEN undefined_column THEN NULL;
    WHEN duplicate_column THEN NULL;
END $$;

-- Drop the old indexes (Postgres doesn't auto-rename indexes when their table is renamed).
DROP INDEX IF EXISTS workload_type_resource_interfaces_version_null_tenant;
DROP INDEX IF EXISTS workload_type_resource_interfaces_version_with_tenant;

CREATE UNIQUE INDEX IF NOT EXISTS workload_type_kartas_version_null_tenant
    ON workload_type_kartas (workload_type_id, version, scope_type)
    WHERE tenant_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS workload_type_kartas_version_with_tenant
    ON workload_type_kartas (workload_type_id, version, scope_type, tenant_id)
    WHERE tenant_id IS NOT NULL;

-- Recreate the view to reference the renamed table and column.
CREATE VIEW external_workloads_type_config_view AS
SELECT
    workload_types.id,
    workload_types.name,
    workload_types.group,
    workload_types.source,
    config.created_at,
    config.created_by,
    config.updated_at,
    config.updated_by,
    COALESCE(
            (SELECT array_to_json(array_agg(row_to_json(rd) ORDER BY rd.created_at))
             FROM workload_type_kartas rd
             WHERE rd.workload_type_id = workload_types.id
               AND (
                 -- First priority: match config's scope
                 (rd.scope_type = config.scope_type AND rd.tenant_id IS NOT DISTINCT FROM config.tenant_id)
                     OR
                     -- Fallback: system scope, only if no config-scope kartas exist
                 (rd.scope_type = 'system' AND rd.tenant_id IS NULL
                     AND NOT EXISTS (
                         SELECT 1 FROM workload_type_kartas k
                         WHERE k.workload_type_id = workload_types.id
                           AND k.scope_type = config.scope_type
                           AND k.tenant_id IS NOT DISTINCT FROM config.tenant_id
                     ))
                 )),
            '[]'::json
    ) AS kartas,
    COALESCE(category.id, system_category.id) AS category_id,
    COALESCE(priority.id, system_priority.id) AS priority_id,
    COALESCE(category.name, system_category.name) AS category_name,
    COALESCE(priority.name, system_priority.name) AS priority_name,
    COALESCE(config.preemptibility, config_system.preemptibility) AS preemptibility,
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

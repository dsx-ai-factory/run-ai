-- Revert Job workload type configuration back to Build category and high priority
UPDATE external_workloads_type_config
SET 
    category_id = (SELECT id FROM workloads_categories WHERE name = 'Build' AND scope_type = 'system' LIMIT 1),
    priority_id = (SELECT id FROM priorities WHERE name = 'high' AND scope_type = 'system' LIMIT 1),
    updated_at = now(),
    updated_by = 'Run.ai'
WHERE workload_type_id = (
    SELECT id FROM workload_types WHERE name = 'Job' LIMIT 1
)
AND scope_type = 'system';

-- Migration to remove legacy Grafana dashboard permission sets
-- The dashboards-overview, dashboards-analytics, dashboards-consumption resource types
-- were removed in 2.25 as part of the Grafana decommission.

-- Remove associations between roles and the legacy dashboard permission sets
DELETE FROM role_to_permission_set
WHERE permission_set_id IN (
    SELECT id FROM permission_sets
    WHERE name IN (
        'Legacy overview dashboard read',
        'Legacy analytics dashboard read',
        'Legacy consumption dashboard read'
    )
);

-- Delete the legacy dashboard permission sets themselves
DELETE FROM permission_sets
WHERE name IN (
    'Legacy overview dashboard read',
    'Legacy analytics dashboard read',
    'Legacy consumption dashboard read'
);

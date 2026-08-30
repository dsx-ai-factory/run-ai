UPDATE permission_sets
SET name = 'NetQ metrics read'
WHERE name = 'Integration metrics read';

UPDATE roles
SET name = 'Metrics scraper'
WHERE name = 'Integration metrics collector';

UPDATE role_permissions
SET resource_type = 'scrapeable-metrics'
WHERE resource_type = 'workload-integration-metrics';
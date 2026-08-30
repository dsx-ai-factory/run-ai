-- this query just to make sure the file is not empty and flyway will not fail
SELECT COUNT(*)
FROM assets_supported_workload_types
WHERE "workspace" = true AND "training" = true AND "distributed" = false;
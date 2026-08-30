CREATE TABLE IF NOT EXISTS last_used_by_workload (
               id UUID NOT NULL PRIMARY KEY,
               asset_id UUID NOT NULL,
               project_id INTEGER NOT NULL,
               tenant_id INTEGER,
               workload_type TEXT NOT NULL,
               workload_id UUID NOT NULL,
               workload_creation_time timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS ByProjectWorloadTypeAndAssetId
    ON last_used_by_workload(project_id, workload_type, asset_id);
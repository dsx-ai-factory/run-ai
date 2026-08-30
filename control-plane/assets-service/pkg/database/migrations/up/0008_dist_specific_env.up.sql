ALTER TABLE workloads ADD COLUMN IF NOT EXISTS dist_specific_env jsonb;
ALTER TABLE workloads_assets ADD COLUMN IF NOT EXISTS dist_replica_type text;
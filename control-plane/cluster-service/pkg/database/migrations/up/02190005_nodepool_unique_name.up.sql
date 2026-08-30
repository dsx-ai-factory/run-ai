CREATE UNIQUE INDEX IF NOT EXISTS np_unique_name ON nodepools (cluster_id, name) WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_parent_id_not_deleted
    ON project(parent_id)
    WHERE deleted_at IS NULL;

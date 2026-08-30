CREATE INDEX IF NOT EXISTS idx_workloads_type ON workloads (type);
CREATE INDEX IF NOT EXISTS idx_workloads_name ON workloads (name);
CREATE INDEX IF NOT EXISTS idx_workloads_project_name ON workloads (project_name);
CREATE INDEX IF NOT EXISTS idx_workloads_creation_time ON workloads (created_at);
CREATE INDEX IF NOT EXISTS idx_workloads_deletion_time ON workloads (deleted_at) where deleted_at is not null;
CREATE INDEX IF NOT EXISTS idx_workloads_submitted_by ON workloads (submitted_by) where submitted_by <> '';
DROP TRIGGER IF EXISTS set_revision_phases_update_time ON revisions;
DROP FUNCTION IF EXISTS update_revision_phases_update_time;

DROP INDEX IF EXISTS idx_phase_create_time_revision;
DROP INDEX IF EXISTS idx_phase_revision;
DROP INDEX IF EXISTS idx_phase_history_tenant_revision;

DROP TABLE IF EXISTS revisions_phase_history;


ALTER TABLE reconciler_versions
DROP COLUMN IF EXISTS updated_by;

ALTER TABLE reconciler_versions
DROP COLUMN IF EXISTS updated_by;

ALTER TABLE reconciler_versions
ADD COLUMN IF NOT EXISTS deleted_at TEXT;

ALTER TABLE effective_policies DROP COLUMN IF EXISTS updated_at;

DROP TRIGGER IF EXISTS set_timestamp ON effective_policies;

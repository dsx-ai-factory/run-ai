ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE effective_policies DROP CONSTRAINT IF EXISTS effective_policies_pkey;
DROP INDEX IF EXISTS effective_policies_pkey;
CREATE UNIQUE INDEX IF NOT EXISTS effective_policies_pkey ON effective_policies (policy_id, deleted_at);

-- when upgrading existing cluster we want to create effective policies soft-deleted records
-- for any deleted policy. this way we get the same database that we would get if we deleted
-- all of them after installation of this version. 
-- 
INSERT INTO effective_policies (policy_id, deleted_at)
    SELECT id, now() from policies
        LEFT JOIN effective_policies ep ON id = ep.policy_id WHERE ep.policy_id is null;

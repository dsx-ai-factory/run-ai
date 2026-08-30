
-- return the index to the same strucutre it was before we updated it in the up migration
DROP INDEX IF EXISTS by_name;
CREATE UNIQUE INDEX IF NOT EXISTS by_name ON policies(tenant_id, workload_type, name, deleted_at);

-- return the index to the same strucutre it was before we updated it in the up migration
DROP INDEX IF EXISTS effective_policies_pkey;
CREATE UNIQUE INDEX IF NOT EXISTS effective_policies_pkey ON effective_policies(policy_id, deleted_at);


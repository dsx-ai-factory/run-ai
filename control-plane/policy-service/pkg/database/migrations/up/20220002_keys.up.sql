
-- delete duplicate policies to avoid isues when re-creating the by_name key
DELETE FROM policies
WHERE policies.ctid NOT IN (
    SELECT min(p.ctid)
    FROM policies p
    GROUP BY tenant_id, workload_type, name
);

-- delete duplicate effective policies
DELETE FROM effective_policies
WHERE effective_policies.ctid NOT IN (
    SELECT min(ep.ctid)
    FROM effective_policies ep
    GROUP BY ep.policy_id
);

-- sanitize the effective policies table to sync with the policies table
DELETE FROM effective_policies WHERE policy_id NOT IN (select id FROM policies);

-- remove the deleted at field from the by_name index
DROP INDEX IF EXISTS by_name;
CREATE UNIQUE INDEX IF NOT EXISTS by_name ON policies(tenant_id, workload_type, name);

-- remove the delete at field from the effective_policies_pkey index
DROP INDEX IF EXISTS effective_policies_pkey;
CREATE UNIQUE INDEX IF NOT EXISTS effective_policies_pkey ON effective_policies(policy_id);


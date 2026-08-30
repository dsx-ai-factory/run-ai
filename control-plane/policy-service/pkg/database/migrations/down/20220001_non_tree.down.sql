
-- drop unique index for type + scope
DROP INDEX IF EXISTS policies_by_type_and_scope;

-- in the up part we remove empty records assuming they are auto-created as part of service runs.
-- for the down part we will remove them again so they will be re-calculated according to the old mechanism.ƒ
DELETE FROM policies WHERE rules_v2 is NULL AND defaults_v2 IS NULL and imposed_assets IS NULL;
DELETE FROM effective_policies WHERE policy_id NOT IN (select id FROM policies);


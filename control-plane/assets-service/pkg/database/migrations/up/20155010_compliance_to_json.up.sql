---- Redo of 0010_compliance_to_json.up.sql in idempotent in case it didn't run
DELETE from policy_compliance WHERE aligned = false;

-- This seems like there is data lost here, how ever this table is a cache, and value will be recalculated
ALTER TABLE policy_compliance DROP COLUMN IF EXISTS reason;
ALTER TABLE policy_compliance ADD COLUMN IF NOT EXISTS reason JSONB;

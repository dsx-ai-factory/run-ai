DELETE from policy_compliance WHERE aligned = false;

ALTER TABLE policy_compliance DROP COLUMN IF EXISTS reason;

ALTER TABLE policy_compliance ADD COLUMN IF NOT EXISTS reason JSONB;

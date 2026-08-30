ALTER TABLE policies DROP COLUMN IF EXISTS "defaults_v2";
ALTER TABLE policies DROP COLUMN IF EXISTS "rules_v2";

ALTER TABLE effective_policies DROP COLUMN IF EXISTS "defaults_v2";
ALTER TABLE effective_policies DROP COLUMN IF EXISTS "rules_v2";


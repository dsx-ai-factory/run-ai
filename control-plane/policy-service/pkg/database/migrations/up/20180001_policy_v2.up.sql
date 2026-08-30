ALTER TABLE policies ADD COLUMN IF NOT EXISTS "defaults_v2" jsonb;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS "rules_v2" jsonb;

ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "defaults_v2" jsonb;
ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "rules_v2" jsonb;


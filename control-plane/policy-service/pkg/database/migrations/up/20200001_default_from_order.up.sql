
ALTER TABLE effective_policies DROP COLUMN IF EXISTS "default_from_order";
ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "default_from_order" jsonb;

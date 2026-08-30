ALTER TABLE policies DROP COLUMN IF EXISTS "imposed_assets";
ALTER TABLE policies ADD COLUMN IF NOT EXISTS "imposed_assets" jsonb;

ALTER TABLE effective_policies DROP COLUMN IF EXISTS "imposed_assets";
ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "imposed_assets" jsonb;


ALTER TABLE policies ADD COLUMN IF NOT EXISTS "imposed_assets" varchar array;

ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "imposed_assets" varchar array;


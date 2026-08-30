
ALTER TABLE effective_policies DROP COLUMN IF EXISTS "validation_error";
ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "validation_error" text;

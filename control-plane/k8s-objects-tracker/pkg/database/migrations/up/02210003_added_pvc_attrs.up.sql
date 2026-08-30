
ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS storage_class_name text;

ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS phase text;

ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS requested_capacity text;
ALTER TABLE pvcs ADD COLUMN IF NOT EXISTS actual_capacity text;

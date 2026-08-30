ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS min_guaranteed_runtime TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_min_guaranteed_runtime TEXT DEFAULT NULL;

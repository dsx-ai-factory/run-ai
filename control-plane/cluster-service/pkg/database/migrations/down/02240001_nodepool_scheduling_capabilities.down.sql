ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_cpu_quota;
ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_quota;
ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_ranks;
ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_adjustable_weights;

ALTER TABLE nodepools DROP COLUMN IF EXISTS status_label_key;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_label_value;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_over_provisioning_ratio;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_gpu_placement_strategy;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_cpu_placement_strategy;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_nodes;
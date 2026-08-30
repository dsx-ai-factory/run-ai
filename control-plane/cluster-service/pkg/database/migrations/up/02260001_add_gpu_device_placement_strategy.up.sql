ALTER TABLE nodepools
    ADD COLUMN IF NOT EXISTS gpu_device_placement_strategy TEXT DEFAULT NULL;

UPDATE nodepools
    SET gpu_device_placement_strategy = gpu_placement_strategy
    WHERE gpu_device_placement_strategy IS NULL;

ALTER TABLE nodepools
    ALTER COLUMN gpu_device_placement_strategy SET NOT NULL,
    ALTER COLUMN gpu_device_placement_strategy SET DEFAULT 'binpack';

ALTER TABLE nodepools
    ADD COLUMN IF NOT EXISTS status_gpu_device_placement_strategy TEXT DEFAULT NULL;

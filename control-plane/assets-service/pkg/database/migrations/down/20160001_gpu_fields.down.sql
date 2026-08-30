

UPDATE assets_compute SET gpu_portion_request = gpu_devices_request WHERE gpu_devices_request = 0 OR gpu_devices_request > 1;

DO $$
BEGIN
ALTER TABLE assets_compute
    RENAME COLUMN gpu_portion_request TO gpu_portion;
EXCEPTION
    WHEN undefined_column THEN NULL;
    WHEN duplicate_column THEN NULL;
END $$;

DO $$
BEGIN
ALTER TABLE assets_compute
    RENAME COLUMN gpu_memory_request TO gpu_memory;
EXCEPTION
    WHEN undefined_column THEN NULL;
    WHEN duplicate_column THEN NULL;
END $$;

ALTER TABLE assets_compute DROP COLUMN IF EXISTS gpu_devices_request;
ALTER TABLE assets_compute DROP COLUMN IF EXISTS gpu_memory_limit;
ALTER TABLE assets_compute DROP COLUMN IF EXISTS gpu_portion_limit;
ALTER TABLE assets_compute DROP COLUMN IF EXISTS gpu_compute_request;
ALTER TABLE assets_compute DROP COLUMN IF EXISTS gpu_compute_limit;
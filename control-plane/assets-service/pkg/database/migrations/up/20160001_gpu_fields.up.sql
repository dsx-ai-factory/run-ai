-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
BEGIN
    ALTER TABLE assets_compute
        RENAME COLUMN gpu_portion TO gpu_portion_request;
EXCEPTION
    WHEN undefined_column THEN NULL;
    WHEN duplicate_column THEN NULL;
END $$;

DO $$
BEGIN
ALTER TABLE assets_compute
    RENAME COLUMN gpu_memory TO gpu_memory_request;
EXCEPTION
    WHEN undefined_column THEN NULL;
    WHEN duplicate_column THEN NULL;
END $$;

ALTER TABLE assets_compute ADD COLUMN IF NOT EXISTS gpu_devices_request INT;
ALTER TABLE assets_compute ADD COLUMN IF NOT EXISTS gpu_memory_limit TEXT;
ALTER TABLE assets_compute ADD COLUMN IF NOT EXISTS gpu_portion_limit NUMERIC;
ALTER TABLE assets_compute ADD COLUMN IF NOT EXISTS gpu_compute_request NUMERIC;
ALTER TABLE assets_compute ADD COLUMN IF NOT EXISTS gpu_compute_limit NUMERIC;

UPDATE assets_compute SET gpu_devices_request = 1                   WHERE (gpu_portion_request > 0 and gpu_portion_request < 1) OR gpu_portion_request is null;
UPDATE assets_compute SET gpu_devices_request = gpu_portion_request WHERE  gpu_portion_request = 0 or gpu_portion_request >= 1;
UPDATE assets_compute SET gpu_portion_request = null                WHERE  gpu_portion_request = 0 or gpu_portion_request > 1;

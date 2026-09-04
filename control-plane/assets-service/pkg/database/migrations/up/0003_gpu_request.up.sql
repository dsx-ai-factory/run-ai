-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE assets_compute ADD COLUMN if not exists gpu_request_type text;
ALTER TABLE assets_compute ADD COLUMN if not exists gpu_portion numeric;
ALTER TABLE assets_compute ADD COLUMN if not exists gpu_memory text;
ALTER TABLE assets_compute ADD COLUMN if not exists mig_profile text;

UPDATE assets_compute SET gpu_portion = (gpu_request->'gpuPortion'->>'portion')::numeric;
UPDATE assets_compute SET gpu_memory = (gpu_request->'gpuMemory'->>'memory')::text;
UPDATE assets_compute SET mig_profile = (gpu_request->'migProfile'->>'profile')::text;

UPDATE assets_compute SET gpu_request_type = CASE
  WHEN gpu_memory IS NOT NULL THEN 'memory'
  WHEN gpu_portion IS NOT NULL THEN 'portion'
  WHEN mig_profile IS NOT NULL THEN 'migProfile'
  ELSE NULL 
END;


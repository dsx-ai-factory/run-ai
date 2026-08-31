-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Add foreign key constraint for environment_asset_id
ALTER TABLE assets_model DROP CONSTRAINT IF EXISTS assets_model_environment_asset_fk;
ALTER TABLE assets_model ADD CONSTRAINT assets_model_environment_asset_fk FOREIGN KEY (environment_asset_id) REFERENCES assets_environment (asset_id) ON DELETE RESTRICT;

-- Add foreign key constraint for compute_asset_id
ALTER TABLE assets_model DROP CONSTRAINT IF EXISTS assets_model_compute_asset_fk;
ALTER TABLE assets_model ADD CONSTRAINT assets_model_compute_asset_fk FOREIGN KEY (compute_asset_id) REFERENCES assets_compute (asset_id) ON DELETE RESTRICT;

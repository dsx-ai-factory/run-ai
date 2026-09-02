-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_config_map DROP COLUMN IF EXISTS default_mode;
ALTER TABLE assets_secret_volume DROP COLUMN IF EXISTS default_mode;
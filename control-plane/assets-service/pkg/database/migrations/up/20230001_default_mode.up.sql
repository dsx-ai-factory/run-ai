-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_config_map ADD COLUMN IF NOT EXISTS default_mode text;
ALTER TABLE assets_secret_volume ADD COLUMN IF NOT EXISTS default_mode text;

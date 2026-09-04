-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


CREATE TABLE IF NOT EXISTS assets_config_map (
    asset_id UUID NOT NULL PRIMARY KEY,
    config_map TEXT NOT NULL,
    mount_path TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS assets_config_map_pkey ON assets_config_map (asset_id);

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS assets_ngc_api_key (
    asset_id UUID NOT NULL PRIMARY KEY,
    ngc_api_key TEXT NOT NULL,
    key_id INTEGER
); 
-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS assets_generic_secret (
                                                 asset_id UUID NOT NULL PRIMARY KEY,
                                                 existing_secret_name TEXT
);

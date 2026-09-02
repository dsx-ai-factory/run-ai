-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_generic_secret DROP COLUMN if exists key_value_pairs;
ALTER TABLE assets_generic_secret DROP COLUMN if exists key_id;
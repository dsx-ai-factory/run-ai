-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_generic_secret ADD COLUMN if not exists key_value_pairs jsonb;
ALTER TABLE assets_generic_secret ADD COLUMN if not exists key_id int4;
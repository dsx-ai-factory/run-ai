-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS access_rules_search_vector_idx;
ALTER TABLE IF EXISTS access_rules DROP COLUMN IF EXISTS search_vector;

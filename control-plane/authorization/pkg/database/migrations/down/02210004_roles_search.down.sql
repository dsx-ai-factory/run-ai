-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS roles_search_vector_idx;
ALTER TABLE IF EXISTS roles DROP COLUMN IF EXISTS search_vector;

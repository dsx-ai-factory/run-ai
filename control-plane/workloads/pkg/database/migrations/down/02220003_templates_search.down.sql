-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS crud_templates_search_vector_idx;
ALTER TABLE IF EXISTS crud_templates DROP COLUMN IF EXISTS search_vector; 
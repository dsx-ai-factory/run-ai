-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_reports_created_by;
DROP INDEX IF EXISTS idx_reports_created_at;
DROP INDEX IF EXISTS idx_reports_status;
DROP INDEX IF EXISTS idx_reports_search_vector;
ALTER TABLE IF EXISTS reports DROP COLUMN IF EXISTS search_vector;

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE reconciler_versions
DROP COLUMN IF EXISTS updated_by;

ALTER TABLE reconciler_versions
DROP COLUMN IF EXISTS updated_by;

ALTER TABLE reconciler_versions
ADD COLUMN IF NOT EXISTS deleted_at TEXT;
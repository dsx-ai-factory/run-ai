-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE reconciler_versions
ADD COLUMN IF NOT EXISTS updated_by TEXT;

ALTER TABLE reconciler_versions
ADD COLUMN IF NOT EXISTS created_by TEXT;

ALTER TABLE reconciler_versions
DROP COLUMN IF EXISTS deleted_at;
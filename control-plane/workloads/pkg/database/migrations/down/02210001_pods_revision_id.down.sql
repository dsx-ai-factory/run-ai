-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS pods_revision_idx;
ALTER TABLE IF EXISTS pods DROP COLUMN IF EXISTS revision_id;

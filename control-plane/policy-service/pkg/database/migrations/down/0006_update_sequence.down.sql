-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TRIGGER IF EXISTS set_update_count ON effective_policies;

DROP INDEX IF EXISTS effective_policies_by_update;

ALTER TABLE effective_policies DROP COLUMN IF EXISTS update_count;
ALTER TABLE effective_policies DROP COLUMN IF EXISTS forced_update;

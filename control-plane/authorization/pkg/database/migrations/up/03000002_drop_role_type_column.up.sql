-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- 03000001 moved every DGXC role that was in use to a tenant-scoped copy and deleted
-- the originals, so nothing reads role_type any more.
--
-- Ship this only in a release whose code has already dropped models.Role.RoleType:
-- the field is a mapped column, so GORM names it in the explicit column list of every
-- SELECT on roles. Any pod still running the old code fails every role read with
-- 'column "role_type" does not exist' once this applies.

ALTER TABLE IF EXISTS roles DROP COLUMN IF EXISTS role_type;

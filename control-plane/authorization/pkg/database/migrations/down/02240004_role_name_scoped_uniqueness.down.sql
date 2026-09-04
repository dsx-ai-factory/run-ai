-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Revert btree_gist approach
-- Drop the exclusion constraint
ALTER TABLE IF EXISTS roles DROP CONSTRAINT IF EXISTS roles_name_scope_exclusion;
-- Drop the scoped unique indexes
DROP INDEX IF EXISTS roles_tenant_name_unique;
DROP INDEX IF EXISTS roles_system_name_unique;
-- Note: Not dropping btree_gist extension as other tables/constraints might depend on it
-- Restore the original global unique constraint on role name
ALTER TABLE IF EXISTS roles ADD CONSTRAINT roles_name_key UNIQUE (name);

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TRIGGER IF EXISTS enforce_tenant_name_alias_cross_uniqueness ON tenants;
DROP FUNCTION IF EXISTS tenant_name_alias_cross_uniqueness();
DROP INDEX IF EXISTS idx_tenants_alias;
ALTER TABLE tenants DROP COLUMN IF EXISTS alias;
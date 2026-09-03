-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Remove category_id from workloads table
ALTER TABLE workloads DROP CONSTRAINT IF EXISTS fk_workloads_category;
ALTER TABLE workloads DROP COLUMN IF EXISTS category_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_categories_tenant;
DROP INDEX IF EXISTS idx_workloads_category;
DROP INDEX IF EXISTS idx_categories_tenant_name_active;
DROP INDEX IF EXISTS idx_external_workloads_type_config_tenant_type;

-- Drop workload type to category mappings table
DROP TABLE IF EXISTS external_workloads_type_config;

-- Drop categories table
DROP TABLE IF EXISTS workloads_categories;

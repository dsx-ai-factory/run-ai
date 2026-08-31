-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Use btree_gist extension with exclusion constraint to enforce proper scoped uniqueness

-- Drop the existing global unique constraint on role name
ALTER TABLE IF EXISTS roles DROP CONSTRAINT IF EXISTS roles_name_key;

-- Drop old COALESCE index if it exists
DROP INDEX IF EXISTS roles_scope_name_unique;
-- Enable btree_gist extension for exclusion constraints
CREATE EXTENSION IF NOT EXISTS btree_gist;
-- Create unique index for system roles: name must be globally unique for system scope
CREATE UNIQUE INDEX IF NOT EXISTS roles_system_name_unique ON roles(name) WHERE scope_type = 'system';
-- Create unique index for tenant roles: (scope_id, name) must be unique per tenant
CREATE UNIQUE INDEX IF NOT EXISTS roles_tenant_name_unique ON roles(scope_id, name) WHERE scope_type = 'tenant' ;
-- Exclusion constraint: prevent same name from existing in both system and tenant scopes
-- This ensures that if a system role has a name, no tenant can use it (and vice versa)
-- We use scope_type WITH <> which means: exclude rows where name is equal AND scope_type is different
ALTER TABLE roles ADD CONSTRAINT roles_name_scope_exclusion
EXCLUDE USING gist (
    name WITH =,
    scope_type WITH <>
);
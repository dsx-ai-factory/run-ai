-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- =============================================================================
-- Migration: Move tenant_settings from backend schema to assets schema
-- =============================================================================
-- The assets service currently reads/writes the 'tenant_settings' table
-- in the 'backend' schema (owned by the old backend service) to track whether
-- preset assets have been cloned to each tenant (key 'workspaces.preset_installed').
-- This creates a cross-schema coupling: the assets service depends on another
-- service's DB schema at runtime.
--
-- This migration creates a local 'tenant_settings' table in the assets schema
-- and copies the relevant data from the backend schema, so the assets service
-- owns its own data. The old backend table is intentionally NOT deleted (at the 
-- end of the migration) to support downgrade scenarios.
--
-- NOTE: The backend schema name 'backend' is hardcoded here because SQL
-- migrations cannot read the Go config (DB_BACKEND_SCHEMA env var).
-- The default/standard value is 'backend'.
-- =============================================================================

-- Step 1: Create the table in the assets schema (idempotent via IF NOT EXISTS)
CREATE TABLE IF NOT EXISTS tenant_settings (
    tenant_id integer NOT NULL,
    key text NOT NULL,
    value text,
    type text,
    PRIMARY KEY (tenant_id, key)
);

-- Step 2: Copy preset-tracking rows from the old backend table, if it exists.
-- Only copies key = 'workspaces.preset_installed' (the only key the assets
-- service uses). ON CONFLICT DO NOTHING ensures idempotency: re-running this
-- migration copies any NEW rows from backend without touching existing ones.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'backend' AND table_name = 'tenant_settings'
    ) THEN
        INSERT INTO tenant_settings (tenant_id, key, value, type)
        SELECT tenant_id, key, value, type
        FROM backend.tenant_settings
        WHERE key = 'workspaces.preset_installed'
        ON CONFLICT (tenant_id, key) DO NOTHING;
    END IF;
END $$;

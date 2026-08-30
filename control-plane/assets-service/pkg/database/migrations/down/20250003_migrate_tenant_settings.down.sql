-- =============================================================================
-- Rollback: Move tenant_settings back to backend schema
-- =============================================================================
-- Ensures the backend schema and tenant_settings table exist (re-creates them
-- if needed), then copies the preset-tracking data back. This guarantees the
-- old backend code has a table to read from after the downgrade, even on
-- environments where the backend schema was never present.
-- Uses DO UPDATE to ensure the backend table gets the latest values, since the
-- assets service may have written newer revisions after the up migration ran.
-- =============================================================================

-- Step 1: Ensure the backend schema and table exist so we always have a
-- destination for the data, regardless of whether the backend was present before.
CREATE SCHEMA IF NOT EXISTS backend;

CREATE TABLE IF NOT EXISTS backend.tenant_settings (
    tenant_id integer NOT NULL,
    key text NOT NULL,
    value text,
    type text,
    PRIMARY KEY (tenant_id, key)
);

-- Step 2: Copy preset-tracking rows back to the backend table.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = current_schema() AND table_name = 'tenant_settings'
    ) THEN
        INSERT INTO backend.tenant_settings (tenant_id, key, value, type)
        SELECT tenant_id, key, value, type
        FROM tenant_settings
        WHERE key = 'workspaces.preset_installed'
        ON CONFLICT (tenant_id, key)
        DO UPDATE SET value = EXCLUDED.value, type = EXCLUDED.type;
    END IF;
END $$;

-- Step 3: Drop the assets-local table.
DROP TABLE IF EXISTS tenant_settings;

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Migration to deprecate the old chatbot-ui environment preset
--
-- Background: chatbot-ui was originally a system-scoped asset (one shared row).
-- Migration 20250001 moved it to the fictive tenant (scope=tenant, tenant_id=2147483647).
-- But it's still a SINGLE row — tenants don't have their own copies.
--
-- Strategy:
-- 1. For each real tenant that has preset assets, clone chatbot-ui as a per-tenant copy
-- 2. Mark those cloned copies as (Legacy)
-- 3. Delete the original from the fictive tenant so new tenants won't get it

DO $$
DECLARE
    chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
    fictive_tenant_id int := 2147483647;
    legacy_description text := '(Legacy) A user interface for interacting with chat-based AI models. Consider using Hugging Face Chat-UI or Open WebUI instead.';
    rec RECORD;
    new_asset_id uuid;
    env_found boolean := false;
    wt_found boolean := false;
    v_image text;
    v_image_pull_policy text;
    v_environment_variables jsonb;
    v_connections jsonb;
    v_workspace boolean;
    v_training boolean;
    v_distributed boolean;
    v_inference boolean;
BEGIN
    -- Only proceed if chatbot-ui exists in the fictive tenant
    IF EXISTS (SELECT 1 FROM assets_metadata WHERE id = chatbot_ui_env_id AND tenant_id = fictive_tenant_id) THEN

        -- Get the environment spec for cloning
        SELECT image, image_pull_policy, environment_variables, connections
        INTO v_image, v_image_pull_policy, v_environment_variables, v_connections
        FROM assets_environment WHERE asset_id = chatbot_ui_env_id;
        env_found := FOUND;

        -- Get workload types for cloning
        SELECT workspace, training, distributed, inference
        INTO v_workspace, v_training, v_distributed, v_inference
        FROM assets_supported_workload_types WHERE asset_id = chatbot_ui_env_id;
        wt_found := FOUND;

        -- For each real tenant that already has preset assets (i.e., was provisioned),
        -- create a per-tenant copy of chatbot-ui marked as Legacy
        FOR rec IN
            SELECT DISTINCT tenant_id
            FROM assets_metadata
            WHERE tenant_id IS NOT NULL
              AND tenant_id != fictive_tenant_id
              AND kind = 'environment'
              -- Only tenants that don't already have their own chatbot-ui
              AND tenant_id NOT IN (
                  SELECT tenant_id FROM assets_metadata
                  WHERE name = 'chatbot-ui' AND kind = 'environment'
                    AND tenant_id IS NOT NULL AND tenant_id != fictive_tenant_id
              )
        LOOP
            new_asset_id := gen_random_uuid();

            -- Clone metadata with Legacy description
            INSERT INTO assets_metadata (id, kind, scope, tenant_id, name, created_by, created_at, updated_by, updated_at, description)
            SELECT new_asset_id, kind, scope, rec.tenant_id, name, created_by, now(), created_by, now(), legacy_description
            FROM assets_metadata
            WHERE id = chatbot_ui_env_id
            ON CONFLICT DO NOTHING;

            -- Clone environment spec
            IF env_found THEN
                INSERT INTO assets_environment (asset_id, image, image_pull_policy, environment_variables, connections)
                VALUES (new_asset_id, v_image, v_image_pull_policy, v_environment_variables, v_connections)
                ON CONFLICT DO NOTHING;
            END IF;

            -- Clone workload supported types
            IF wt_found THEN
                INSERT INTO assets_supported_workload_types (asset_id, workspace, training, distributed, inference)
                VALUES (new_asset_id, v_workspace, v_training, v_distributed, v_inference)
                ON CONFLICT DO NOTHING;
            END IF;
        END LOOP;

        -- Now delete the original from the fictive tenant
        DELETE FROM assets_supported_workload_types WHERE asset_id = chatbot_ui_env_id;
        DELETE FROM assets_environment WHERE asset_id = chatbot_ui_env_id;
        DELETE FROM assets_metadata WHERE id = chatbot_ui_env_id;
    END IF;

    -- Also mark any pre-existing per-tenant copies as Legacy (edge case: if some tenant
    -- already had a cloned copy from preset provisioning)
    UPDATE assets_metadata
    SET description = legacy_description
    WHERE name = 'chatbot-ui' AND kind = 'environment'
      AND tenant_id IS NOT NULL AND tenant_id != fictive_tenant_id;
END $$;

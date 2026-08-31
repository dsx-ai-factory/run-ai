-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Migration to add Chat-UI and OpenWebUI environment presets
-- Chat-UI: Two versions (host-based and path-based routing)
-- OpenWebUI: One version (host-based routing only)

DO $$
DECLARE
    CREATED_BY text := 'Run:ai';
    CREATED_AT timestamp := now();

    -- UUIDs for new environment presets
    uuid_environment_chatui_host uuid := '4ca2338b-1f43-4123-b349-57d6ad898542';
    uuid_environment_chatui_path uuid := '01a576e4-a02f-4866-ae90-912265ffee0a';
    uuid_environment_openwebui uuid := 'd0e72d44-c5c8-4a2b-b8b8-9d062de47b7c';

    -- Description text for environments
    description_chatui_host text := 'Hugging Face Chat-UI with host-based routing - A clean, modern chat interface using unique subdomains for each workspace. Requires wildcard DNS support.';
    description_chatui_path text := 'Hugging Face Chat-UI with path-based routing - A chat interface using path-based URLs, suitable for environments without wildcard DNS or behind corporate proxies.';
    description_openwebui text := 'Open WebUI (host-based routing) - A feature-rich web interface for LLM interactions with RAG support, model management, and user authentication. Supports OpenAI API compatible endpoints.';

    -- Docker images from running workloads
    chatui_image text := 'ghcr.io/run-ai/chat-ui-db:0.20';
    openwebui_image text := 'ghcr.io/open-webui/open-webui:v0.6.42';

    -- Common settings
    image_pull_policy text := 'IfNotPresent';
    uid_gid_source text := 'fromTheImage';

BEGIN
    BEGIN
        -- Insert metadata for all three environments
        INSERT INTO assets_metadata ("id", "kind", "scope", "tenant_id", "name", "created_by", "created_at", "updated_by", "updated_at", "description") VALUES
            (uuid_environment_chatui_host, 'environment', 'tenant', '2147483647', 'huggingface-chat-ui', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, description_chatui_host),
            (uuid_environment_chatui_path, 'environment', 'tenant', '2147483647', 'huggingface-chat-ui-path', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, description_chatui_path),
            (uuid_environment_openwebui, 'environment', 'tenant', '2147483647', 'openwebui', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, description_openwebui);

        -- Insert Chat-UI HOST-BASED environment
        INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "connections", "environment_variables", "uid_gid_source", "routing_method") VALUES
            (uuid_environment_chatui_host, 
             chatui_image,
             image_pull_policy,
             '[{"name": "Chat UI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 3000, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]',
             '[{"name": "OPENAI_BASE_URL", "value": "", "description": "Base URL for single model configuration. Example: http://nim-service.runai-project.svc.cluster.local/v1"}, {"name": "DEFAULT_MAX_TOKENS", "value": "4096", "description": "Fallback max_tokens applied to all models unless overridden in MODELS."}, {"name": "PUBLIC_ORIGIN", "value": "", "description": "Full public URL of this workspace. Example: https://project-workload.domain.com/"}, {"name": "APP_BASE", "value": "/"}]',
             uid_gid_source,
             'host-base');

        -- Insert Chat-UI PATH-BASED environment
        INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "connections", "environment_variables", "uid_gid_source", "routing_method") VALUES
            (uuid_environment_chatui_path,
             chatui_image,
             image_pull_policy,
             '[{"name": "Chat UI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 3000, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]',
             '[{"name": "OPENAI_BASE_URL", "value": "", "description": "Base URL for single model configuration. Example: http://nim-service.runai-project.svc.cluster.local/v1"}, {"name": "DEFAULT_MAX_TOKENS", "value": "4096", "description": "Fallback max_tokens applied to all models unless overridden in MODELS."}, {"name": "APP_BASE", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}"}]',
             uid_gid_source,
             'path-base');

        -- Insert OpenWebUI HOST-BASED environment
        INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "connections", "environment_variables", "uid_gid_source", "routing_method") VALUES
            (uuid_environment_openwebui,
             openwebui_image,
             image_pull_policy,
             '[{"name": "Open WebUI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 8080, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]',
             '[{"name": "OPENAI_API_BASE_URL", "value": "", "description": "Base URL for your LLM inference service. Example: http://nim-service.runai-project.svc.cluster.local/v1"}, {"name": "DATABASE_URL", "value": "sqlite:////tmp/webui.db"}, {"name": "DATA_DIR", "value": "/tmp"}, {"name": "ENABLE_OLLAMA_API", "value": "false"}, {"name": "ENABLE_OPENAI_API", "value": "true"}, {"name": "ENABLE_RAG", "value": "false"}, {"name": "HF_HOME", "value": "/tmp/cache"}, {"name": "WEBUI_ALLOW_REGISTRATION", "value": "true"}, {"name": "WEBUI_AUTH", "value": "false"}, {"name": "WEBUI_SECRET_KEY", "value": "sk-fake-key"}]',
             uid_gid_source,
             'host-base');

        -- Set supported workload types for all environments (workspace and training)
        -- Note: dist_framework must be NULL (not 'false') to avoid frontend icon issues
        INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "dist_framework", "inference") VALUES
            (uuid_environment_chatui_host, TRUE, TRUE, FALSE, NULL, FALSE),
            (uuid_environment_chatui_path, TRUE, TRUE, FALSE, NULL, FALSE),
            (uuid_environment_openwebui, TRUE, FALSE, FALSE, NULL, FALSE);  -- OpenWebUI only for workspace

    EXCEPTION
        -- Ensure migration is idempotent
        WHEN unique_violation THEN
            RAISE NOTICE 'Environment assets already exist, skipping';
    END;
END $$;


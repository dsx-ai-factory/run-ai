-- Rollback: restore chatbot-ui to the fictive tenant and remove per-tenant Legacy copies

DO $$
DECLARE
    CREATED_BY text := 'Run:ai';
    CREATED_AT timestamp := now();
    IMAGE_PULL_POLICY text := 'IfNotPresent';

    chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
    fictive_tenant_id int := 2147483647;
    original_description text := 'A user interface for interacting with chat-based AI models, often used for testing and deploying chatbot applications';
    chatbot_ui_env_environment_variables jsonb := '[{"name": "RUNAI_MODEL_NAME", "value": "MODEL_NAME_HERE"}, {"name": "RUNAI_MODEL_BASE_URL", "value": "http://WORKLOAD.runai-PROJECT.svc"}, {"name": "RUNAI_MODEL_MAX_LENGTH", "value": "4096"}, {"name": "RUNAI_MODEL_TOKEN_LIMIT", "value": "2048"}, {"name": "PATH_PREFIX", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}"}]';
    chatbot_ui_env_connections jsonb := '[{"name":"Chatbot UI","internalToolInfo":{"toolType":"chatbot-ui","containerPort":3000,"connectionType":"ExternalUrl","externalUrlInfo":{}}}]';
    rec RECORD;
BEGIN
    -- Delete the per-tenant Legacy copies that were created by the up migration
    FOR rec IN
        SELECT id FROM assets_metadata
        WHERE name = 'chatbot-ui' AND kind = 'environment'
          AND tenant_id IS NOT NULL AND tenant_id != fictive_tenant_id
    LOOP
        DELETE FROM assets_supported_workload_types WHERE asset_id = rec.id;
        DELETE FROM assets_environment WHERE asset_id = rec.id;
        DELETE FROM assets_metadata WHERE id = rec.id;
    END LOOP;

    -- Restore chatbot-ui in the fictive tenant
    BEGIN
        INSERT INTO assets_metadata ("id", "kind", "scope", "tenant_id", "name", "created_by", "created_at", "updated_by", "updated_at", "description")
        VALUES (chatbot_ui_env_id, 'environment', 'tenant', fictive_tenant_id, 'chatbot-ui', CREATED_BY, CREATED_AT, CREATED_BY, CREATED_AT, original_description);

        INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "environment_variables", "connections")
        VALUES (chatbot_ui_env_id, 'gcr.io/run-ai-demo/llm:app', IMAGE_PULL_POLICY, chatbot_ui_env_environment_variables, chatbot_ui_env_connections);

        INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference")
        VALUES (chatbot_ui_env_id, TRUE, TRUE, FALSE, TRUE);
    EXCEPTION
        WHEN unique_violation THEN
            RAISE NOTICE 'chatbot-ui already exists in fictive tenant, skipping restore';
    END;
END $$;

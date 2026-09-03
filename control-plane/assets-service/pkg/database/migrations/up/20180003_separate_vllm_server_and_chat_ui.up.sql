-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        CREATED_OR_UPDATED_AT timestamp := now();
        CREATED_BY text := 'Run:ai';
        IMAGE_PULL_POLICY text := 'IfNotPresent';

        --     old environments
        uuid_environment_falcon_7b uuid := '514b77df-0869-468d-94ce-5a2c7b5dc1bb';
        uuid_environment_falcon_40b uuid := 'edafbeda-9bf5-4b8f-8483-40f8d006b81b';
        uuid_environment_llama_2_7b uuid := 'eed318a0-4e7c-4ccf-8af3-79d28e733f50';
        uuid_environment_llama_2_13b uuid := 'bf571c89-e258-4954-9c09-93c59f36ace5';

        --     models
        uuid_model_falcon_7b uuid := 'c6accd0c-3652-4975-ba11-7c9e3e05c7dd';
        uuid_model_falcon_40b uuid := 'b074bd87-6c31-41fa-89f2-16b470d46d29';
        uuid_model_llama_2_7b uuid := '45ab1c3e-df3b-4b8e-819e-b477fac1302f';
        uuid_model_llama_2_13b uuid := '5be388f5-2420-4aaa-b63d-73ee65b69211';
        uuid_model_gpt2 uuid := '608af4f4-1de5-490a-ab14-3fff85111e52';


        --    new environments
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
        gpt2_env_id uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';

        --    new assets_environment variables
        llm_server_env_environment_variables jsonb := '[{"name": "RUNAI_MODEL", "value": "MODEL_NAME_OR_PATH_HERE"}, {"name": "RUNAI_MODEL_NAME", "value": "MODEL_NAME_HERE"}]';
        chatbot_ui_env_environment_variables jsonb := '[{"name": "RUNAI_MODEL_NAME", "value": "MODEL_NAME_HERE"}, {"name": "RUNAI_MODEL_BASE_URL", "value": "http://WORKLOAD.runai-PROJECT.svc"}, {"name": "RUNAI_MODEL_MAX_LENGTH", "value": "4096"}, {"name": "RUNAI_MODEL_TOKEN_LIMIT", "value": "2048"}, {"name": "PATH_PREFIX", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}"}]';
        gpt2_env_environment_variables jsonb := '[{"name": "PATH_PREFIX", "value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}"}]';

        --     new connections
        llm_server_env_connections jsonb := '[{"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 8000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]';
        chatbot_ui_env_connections jsonb := '[{"name":"Chatbot UI","internalToolInfo":{"toolType":"chatbot-ui","containerPort":3000,"connectionType":"ExternalUrl","externalUrlInfo":{}}}]';
        gpt2_env_connections jsonb := '[{"name": "Chatbot UI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 3000, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}, {"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 3000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]';

        THUMBNAIL_GPT2 bytea := '';
        LICENSE_GPT2 bytea := 'runai internal';


    BEGIN
        BEGIN
            ALTER TABLE assets_model ALTER COLUMN "compute_asset_id" DROP NOT NULL;

            --         create new assets_environment

            INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "environment_variables", "connections")
            VALUES(llm_server_env_id, 'gcr.io/run-ai-demo/llm', IMAGE_PULL_POLICY, llm_server_env_environment_variables, llm_server_env_connections);

            INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "environment_variables", "connections")
            VALUES(chatbot_ui_env_id, 'gcr.io/run-ai-demo/llm:app', IMAGE_PULL_POLICY, chatbot_ui_env_environment_variables, chatbot_ui_env_connections);

            INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "environment_variables", "connections")
            VALUES(gpt2_env_id, 'gcr.io/run-ai-demo/quickstart-inference:gpt2-cpu', IMAGE_PULL_POLICY, gpt2_env_environment_variables, gpt2_env_connections);


            INSERT INTO assets_metadata ("id", "kind", "scope", "name", "created_by", "created_at", "updated_by", "updated_at")
            VALUES(llm_server_env_id, 'environment', 'system', 'llm-server', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT);

            INSERT INTO assets_metadata ("id", "kind", "scope", "name", "created_by", "created_at", "updated_by", "updated_at")
            VALUES(chatbot_ui_env_id, 'environment', 'system', 'chatbot-ui', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT);

            INSERT INTO assets_metadata ("id", "kind", "scope", "name", "created_by", "created_at", "updated_by", "updated_at")
            VALUES(gpt2_env_id, 'environment', 'system', 'gpt2', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT);


            INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference")
            VALUES(llm_server_env_id, FALSE, FALSE, FALSE, TRUE);

            INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference")
            VALUES(chatbot_ui_env_id, TRUE, TRUE, FALSE, TRUE);

            INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference")
            VALUES(gpt2_env_id, FALSE, FALSE, FALSE, TRUE);

            --        update existing models to use new assets_environment id and clean license
            UPDATE
                assets_model
            SET
                "environment_asset_id" = llm_server_env_id,
                "license" = ''
            WHERE
                asset_id IN(uuid_model_falcon_7b, uuid_model_falcon_40b, uuid_model_llama_2_7b, uuid_model_llama_2_13b);

            --         delete old environment assets (not in use)
            DELETE FROM assets_environment
            WHERE asset_id IN(uuid_environment_falcon_7b, uuid_environment_falcon_40b, uuid_environment_llama_2_7b, uuid_environment_llama_2_13b);

            DELETE FROM assets_metadata
            WHERE id IN(uuid_environment_falcon_7b, uuid_environment_falcon_40b, uuid_environment_llama_2_7b, uuid_environment_llama_2_13b);

            DELETE FROM assets_supported_workload_types
            WHERE asset_id IN(uuid_environment_falcon_7b, uuid_environment_falcon_40b, uuid_environment_llama_2_7b, uuid_environment_llama_2_13b);

            --         finally add gpt2 model to models after new environment is created
            INSERT INTO assets_model ("asset_id", "title", "description", "thumbnail", "huggingface_id", "license", "environment_asset_id") VALUES
                (uuid_model_gpt2, 'GPT-2', 'OpenAI''s GPT2 Large is the 774M parameter version of GPT-2, a transformer-based language model created and released by OpenAI. The model is a pre-trained model on English language using a causal language modeling (CLM) objective.', THUMBNAIL_GPT2, 'openai-community/gpt2', LICENSE_GPT2, gpt2_env_id);

            INSERT INTO assets_metadata ("id", "kind", "scope", "name", "created_by", "created_at", "updated_by", "updated_at")
            VALUES(uuid_model_gpt2, 'model', 'system', 'GPT-2', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT);

        END;
    END
$$;

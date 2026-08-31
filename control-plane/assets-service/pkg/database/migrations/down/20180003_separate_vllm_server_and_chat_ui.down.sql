-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        --     old environments
        CREATED_OR_UPDATED_AT timestamp := now();
        CREATED_BY text := 'Run:ai';

        uuid_environment_falcon_7b uuid := '514b77df-0869-468d-94ce-5a2c7b5dc1bb';
        uuid_environment_falcon_40b uuid := 'edafbeda-9bf5-4b8f-8483-40f8d006b81b';
        uuid_environment_llama_2_7b uuid := 'eed318a0-4e7c-4ccf-8af3-79d28e733f50';
        uuid_environment_llama_2_13b uuid := 'bf571c89-e258-4954-9c09-93c59f36ace5';

        LICENSE_META text := 'This inference workload uses <a target="_blank" href="https://ai.meta.com/resources/models-and-libraries/llama-downloads/">Llama2</a>, which is a gated model.<div>Your use of this inference workload is conditioned on continued compliance with Meta''s <a target="_blank" href="https://ai.meta.com/llama/use-policy/">Llama2 Use Policy</a>. You will indemnify and hold harmless Run.ai from and against any claim by any third party arising out of or related to your use of this workload or violation of the aforementioned Use Policy. Run.ai disclaims all warranties in respect of the workload. Your use of this workload and any output thereof is at your own risk. Run.ai has no liability with respect to this workload, or your use of any output thereof.</div><div>If you do not agree with the foregoing disclaimers and limitations then do not use this workload.</div>';

        DOCKER_IMAGE text := 'gcr.io/run-ai-demo/quickstart-inference';
        ENV_VARS jsonb := '[{"name": "PATH_PREFIX","value": "/${RUNAI_PROJECT}/${RUNAI_JOB_NAME}"}]';
        CONNECTIONS jsonb := '[{"name": "Chatbot UI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 3000, "connectionType": "ExternalUrl", "externalUrlInfo": {}}},{"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 3000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]';
        PROBES_JSON jsonb := '{"readiness": {"handler": {"httpGet": {"port": 3001, "path": "/ready"}}}}';

        --     old models
        uuid_model_falcon_7b uuid := 'c6accd0c-3652-4975-ba11-7c9e3e05c7dd';
        uuid_model_falcon_40b uuid := 'b074bd87-6c31-41fa-89f2-16b470d46d29';
        uuid_model_llama_2_7b uuid := '45ab1c3e-df3b-4b8e-819e-b477fac1302f';
        uuid_model_llama_2_13b uuid := '5be388f5-2420-4aaa-b63d-73ee65b69211';

        --     new models
        uuid_model_gpt2 uuid := '608af4f4-1de5-490a-ab14-3fff85111e52';

        --     new environments
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
        gpt2_env_id uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';

    BEGIN
        BEGIN
            --         create old assets_environment
            INSERT INTO assets_environment ("asset_id", "image", "image_pull_policy", "environment_variables", "connections", probes) VALUES
            (uuid_environment_falcon_7b, DOCKER_IMAGE || ':falcon-7b-instruct', 'IfNotPresent', ENV_VARS, CONNECTIONS, PROBES_JSON),
            (uuid_environment_falcon_40b, DOCKER_IMAGE || ':falcon-40b-instruct', 'IfNotPresent', ENV_VARS, CONNECTIONS, PROBES_JSON),
            (uuid_environment_llama_2_7b, DOCKER_IMAGE || ':Llama-2-7b-chat-hf', 'IfNotPresent', ENV_VARS, CONNECTIONS, PROBES_JSON),
            (uuid_environment_llama_2_13b, DOCKER_IMAGE || ':Llama-2-13b-chat-hf', 'IfNotPresent', ENV_VARS, CONNECTIONS, PROBES_JSON);


            INSERT INTO assets_metadata ("id", "kind", "scope", "name", "created_by", "created_at", "updated_by", "updated_at") VALUES
            (uuid_environment_falcon_7b, 'environment', 'system', 'falcon-7b-instruct', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT),
            (uuid_environment_falcon_40b, 'environment', 'system', 'falcon-40b-instruct', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT),
            (uuid_environment_llama_2_7b, 'environment', 'system', 'llama-2-7b-chat-hf', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT),
            (uuid_environment_llama_2_13b, 'environment', 'system', 'llama-2-13b-chat-hf', CREATED_BY, CREATED_OR_UPDATED_AT, CREATED_BY, CREATED_OR_UPDATED_AT);


            INSERT INTO assets_supported_workload_types ("asset_id", "workspace", "training", "distributed", "inference") VALUES
            (uuid_environment_falcon_7b, FALSE, FALSE, FALSE, TRUE),
            (uuid_environment_falcon_40b, FALSE, FALSE, FALSE, TRUE),
            (uuid_environment_llama_2_7b, FALSE, FALSE, FALSE, TRUE),
            (uuid_environment_llama_2_13b, FALSE, FALSE, FALSE, TRUE);

            --        update existing models to use old assets_environment id and revert license
            UPDATE
                assets_model
            SET
                "environment_asset_id" = uuid_environment_falcon_7b,
                "license" = ''
            WHERE
                asset_id = uuid_model_falcon_7b;

            --           falcon_40b
            UPDATE
                assets_model
            SET
                "environment_asset_id" = uuid_environment_falcon_40b,
                "license" = ''
            WHERE
                asset_id = uuid_model_falcon_40b;

            --           llama_2_13b
            UPDATE
                assets_model
            SET
                "environment_asset_id" = uuid_environment_llama_2_13b,
                "license" = LICENSE_META
            WHERE
                asset_id = uuid_model_llama_2_13b;

            --            llama_2_7b
            UPDATE
                assets_model
            SET
                "environment_asset_id" = uuid_environment_llama_2_7b,
                "license" = LICENSE_META
            WHERE
                asset_id = uuid_model_llama_2_7b;

            --         delete gpt2 model
            delete from assets_model where asset_id = uuid_model_gpt2;
            delete from assets_metadata where id = uuid_model_gpt2;

            --         delete new assets_environment
            delete from assets_environment where asset_id
            in(llm_server_env_id, chatbot_ui_env_id, gpt2_env_id);

            delete from assets_metadata where id
            in(llm_server_env_id, chatbot_ui_env_id, gpt2_env_id);

            delete from assets_supported_workload_types where asset_id
            in(llm_server_env_id, chatbot_ui_env_id, gpt2_env_id);



        END;
    END
$$;

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
        gpt2_env_id       uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';

        llm_server_image text := 'runai.jfrog.io/core-llm/runai-vllm:v0.5.5-0.5.0';
        chatbot_ui_image text := 'runai.jfrog.io/core-llm/llm-app';
        gpt2_image       text := 'ghcr.io/huggingface/text-generation-inference:3.3.6-intel-cpu';
        gpt2_command     text := 'text-generation-launcher';
        gpt2_args        text := '--model-id gpt2 --port 3000';
        gpt2_environment_variables jsonb := '[]';
        gpt2_connections jsonb := '[{"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 3000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]';
    BEGIN
        BEGIN
            UPDATE assets_environment
            SET image = llm_server_image
            WHERE asset_id = llm_server_env_id;

            UPDATE assets_environment
            SET image = chatbot_ui_image
            WHERE asset_id = chatbot_ui_env_id;

            UPDATE assets_environment
            SET image = gpt2_image,
                command = gpt2_command,
                args = gpt2_args,
                environment_variables = gpt2_environment_variables,
                connections = gpt2_connections
            WHERE asset_id = gpt2_env_id;
        END;
    END
$$;

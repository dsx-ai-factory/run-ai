-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        llm_server_image text := 'docker.io/vllm/vllm-openai:v0.6.4';
        llm_server_command text := 'bash';
        llm_server_args text := '-c "if [ -n \"$RUNAI_MODEL_NAME\" ]; then exec python3 -m vllm.entrypoints.openai.api_server --model \"$RUNAI_MODEL\" --served-model-name \"$RUNAI_MODEL_NAME\" --host 0.0.0.0 --port 8000; else exec python3 -m vllm.entrypoints.openai.api_server --model \"$RUNAI_MODEL\" --host 0.0.0.0 --port 8000; fi"';
    BEGIN
        BEGIN
            UPDATE assets_environment
            SET image = llm_server_image,
                command = llm_server_command,
                args = llm_server_args
            WHERE asset_id = llm_server_env_id;
        END;
    END
$$;

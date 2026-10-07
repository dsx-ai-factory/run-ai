-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
DECLARE
    old_images text[] := ARRAY[
        'gcr.io/run-ai-demo/' || 'llm',
        'runai.jfrog.io/core-llm/' || 'runai-vllm:v0.5.5-0.5.0',
        'runai.jfrog.io/core-llm/' || 'runai-vllm:v0.6.4-0.10.0'
    ];
    replacement_image text := 'vllm/vllm-openai:v0.6.4';
    replacement_command text := 'bash';
    replacement_args text := '-c "if [ -n \"$RUNAI_MODEL_NAME\" ]; then exec python3 -m vllm.entrypoints.openai.api_server --model \"$RUNAI_MODEL\" --served-model-name \"$RUNAI_MODEL_NAME\" --host 0.0.0.0 --port 8000; else exec python3 -m vllm.entrypoints.openai.api_server --model \"$RUNAI_MODEL\" --host 0.0.0.0 --port 8000; fi"';
BEGIN
    UPDATE assets_environment AS environment
    SET image = replacement_image,
        command = CASE
            WHEN NULLIF(BTRIM(environment.command), '') IS NULL
             AND NULLIF(BTRIM(environment.args), '') IS NULL
            THEN replacement_command
            ELSE environment.command
        END,
        args = CASE
            WHEN NULLIF(BTRIM(environment.command), '') IS NULL
             AND NULLIF(BTRIM(environment.args), '') IS NULL
            THEN replacement_args
            ELSE environment.args
        END
    FROM assets_metadata AS metadata
    WHERE environment.asset_id = metadata.id
      AND metadata.kind = 'environment'
      AND metadata.name = 'llm-server'
      AND environment.image = ANY(old_images);
END $$;

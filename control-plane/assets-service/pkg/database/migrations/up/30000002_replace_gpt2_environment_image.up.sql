-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
DECLARE
    old_images text[] := ARRAY[
        'gcr.io/run-ai-demo/' || 'quickstart-inference:gpt2-cpu',
        'runai.jfrog.io/core-llm/' || 'quickstart-inference:gpt2-cpu'
    ];
    replacement_image text := 'ghcr.io/huggingface/text-generation-inference:3.3.6-intel-cpu';
    replacement_command text := 'text-generation-launcher';
    replacement_args text := '--model-id gpt2 --port 3000';
    legacy_connections jsonb := '[{"name": "Chatbot UI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 3000, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}, {"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 3000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]';
    replacement_connections jsonb := '[{"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 3000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]';
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
        END,
        connections = CASE
            WHEN environment.connections IS NULL
              OR environment.connections = '[]'::jsonb
              OR environment.connections = legacy_connections
            THEN replacement_connections
            ELSE environment.connections
        END
    FROM assets_metadata AS metadata
    WHERE environment.asset_id = metadata.id
      AND metadata.kind = 'environment'
      AND metadata.name = 'gpt2'
      AND environment.image = ANY(old_images);
END $$;

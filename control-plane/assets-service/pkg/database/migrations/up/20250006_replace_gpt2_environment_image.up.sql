-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
DECLARE
    replacement_image text := 'ghcr.io/huggingface/text-generation-inference:3.3.6-intel-cpu';
    replacement_command text := 'text-generation-launcher';
    replacement_args text := '--model-id gpt2 --port 3000';
    replacement_connections jsonb := '[{"name": "ServingPort", "internalToolInfo": {"toolType": "custom", "containerPort": 3000, "connectionType": "ServingPort", "servingPortInfo": {"protocol": "http"}}}]';
BEGIN
    UPDATE assets_environment AS environment
    SET image = replacement_image,
        command = replacement_command,
        args = replacement_args,
        environment_variables = '[]'::jsonb,
        connections = replacement_connections
    FROM assets_metadata AS metadata
    WHERE environment.asset_id = metadata.id
      AND metadata.kind = 'environment'
      AND metadata.name = 'gpt2'
      AND environment.image = 'runai.jfrog.io/core-llm/' || 'quickstart-inference:gpt2-cpu';
END $$;

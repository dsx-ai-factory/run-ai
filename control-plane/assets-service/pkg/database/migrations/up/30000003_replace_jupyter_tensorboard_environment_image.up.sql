-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
DECLARE
    preset_asset_id uuid := '6d9a2313-f447-40af-bd61-56bafa249e52';
    old_image text := 'gcr.io/run-ai-demo/' || 'jupyter-tensorboard';
    replacement_image text := 'nvcr.io/nvidia/pytorch:26.09-py3';
    replacement_command text := 'bash';
    replacement_args text := '-c "tensorboard --logdir \"${LOG_DIR:-./}\" --bind_all & exec jupyter lab --ip=0.0.0.0 --allow-root --ServerApp.token="';
    replacement_environment_variables jsonb := '[{"name": "LOG_DIR", "value": "./"}]';
    replacement_connections jsonb := '[{"name": "Jupyter", "internalToolInfo": {"toolType": "jupyter-notebook", "containerPort": 8888, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}, {"name": "TensorBoard", "internalToolInfo": {"toolType": "tensorboard", "containerPort": 6006, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]';
BEGIN
    INSERT INTO assets_environment (
        asset_id, command, args, environment_variables, image,
        image_pull_policy, connections, capabilities, shared_memory, uid_gid_source
    )
    SELECT
        metadata.id, replacement_command, replacement_args,
        replacement_environment_variables, replacement_image,
        'IfNotPresent', replacement_connections, '[]'::jsonb, FALSE, 'fromTheImage'
    FROM assets_metadata AS metadata
    WHERE metadata.id = preset_asset_id
      AND metadata.kind = 'environment'
      AND metadata.name = 'jupyter-tensorboard'
      AND NOT EXISTS (
          SELECT 1 FROM assets_environment AS environment
          WHERE environment.asset_id = metadata.id
      );

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
      AND metadata.name = 'jupyter-tensorboard'
      AND environment.image = old_image;
END $$;

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
DECLARE
    old_image text := 'gcr.io/run-ai-demo/' || 'jupyter-tensorboard';
    replacement_image text := 'nvcr.io/nvidia/pytorch:26.09-py3';
    replacement_command text := 'bash';
    replacement_args text := '-c "tensorboard --logdir \"${LOG_DIR:-./}\" --bind_all & exec jupyter lab --ip=0.0.0.0 --allow-root --ServerApp.token="';
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
      AND metadata.name = 'jupyter-tensorboard'
      AND environment.image = old_image;
END $$;

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        llm_server_image text := 'runai.jfrog.io/core-llm/runai-vllm:v0.5.5-0.5.0';
    BEGIN
        BEGIN
            UPDATE assets_environment
            SET image = llm_server_image
            WHERE asset_id = llm_server_env_id;
        END;
    END
$$;

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        llm_server_env_id uuid := 'a50ef834-4008-4a3a-9ac8-bfc9c177f5e4';
        chatbot_ui_env_id uuid := '715a51c3-93ce-4838-a6c6-4d248e6b9ba0';
        gpt2_env_id       uuid := 'dbf64cad-5652-4698-b24c-b7193230acbe';
    BEGIN
        BEGIN
            UPDATE assets_metadata
            SET scope = 'tenant', tenant_id = 2147483647
            WHERE id IN (llm_server_env_id, chatbot_ui_env_id, gpt2_env_id);
        END;
    END
$$;

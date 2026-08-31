-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        PREV_CONNECTIONS jsonb := '[{"name": "Chatbot UI", "internalToolInfo": {"toolType": "chatbot-ui", "containerPort": 3000, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]';

        uuid_environment_falcon_7b uuid := '514b77df-0869-468d-94ce-5a2c7b5dc1bb';
        uuid_environment_falcon_40b uuid := 'edafbeda-9bf5-4b8f-8483-40f8d006b81b';
        uuid_environment_llama_2_7b uuid := 'eed318a0-4e7c-4ccf-8af3-79d28e733f50';
        uuid_environment_llama_2_13b uuid := 'bf571c89-e258-4954-9c09-93c59f36ace5';
    BEGIN
        BEGIN
            UPDATE assets_environment
            SET connections = PREV_CONNECTIONS
            WHERE asset_id in (uuid_environment_falcon_7b, uuid_environment_falcon_40b, uuid_environment_llama_2_7b, uuid_environment_llama_2_13b);
        END;
    END $$;

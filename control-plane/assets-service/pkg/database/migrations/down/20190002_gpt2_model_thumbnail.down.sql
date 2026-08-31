-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
    DECLARE
        uuid_model_gpt2 uuid := '608af4f4-1de5-490a-ab14-3fff85111e52';

        THUMBNAIL_GPT2 bytea := '';
    BEGIN
        BEGIN
            UPDATE assets_model
            SET thumbnail = THUMBNAIL_GPT2
            WHERE asset_id = uuid_model_gpt2;
        END;
    END
$$;

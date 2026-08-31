-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

---- Redo of 0009_workload_template.up.sql in idempotent in case it didn't run
DROP TRIGGER IF EXISTS validate_asset_name ON assets_metadata;

update assets_metadata set kind = 'workload-template' where kind = 'workspace-template';

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER validate_asset_name
            BEFORE INSERT OR UPDATE
            ON assets_metadata
            FOR EACH ROW
            EXECUTE FUNCTION asset_name_validation();
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;
END $$;
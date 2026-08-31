-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_git ADD COLUMN IF NOT EXISTS credential_id uuid;

DO $$
BEGIN
    -- Check if the foreign key constraint 'fk_assets_git_credential_id' already exists
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_assets_git_credential_id'
        AND conrelid = 'assets_git'::regclass -- Table OID for 'assets_git'
    ) THEN
        -- If it doesn't exist, add the constraint
        ALTER TABLE assets_git ADD CONSTRAINT fk_assets_git_credential_id FOREIGN KEY (credential_id) REFERENCES assets_metadata(id);
    END IF;
END
$$;
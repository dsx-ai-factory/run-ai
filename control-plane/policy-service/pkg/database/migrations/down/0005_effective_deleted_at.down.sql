-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- since deleted_at is currently a part of an index, we must remove the rows that have it set before changing the index.
-- otherwise, we'll fail due to multiple rows with the same policy_id, which is supposed to be unique.
DO $$
BEGIN
    IF EXISTS (
        SELECT column_name
        FROM information_schema.columns
        WHERE table_name = 'effective_policies'
        AND column_name = 'deleted_at'
    ) THEN
        DELETE FROM effective_policies WHERE deleted_at IS NOT NULL;
    END IF;
END $$;

DROP INDEX IF EXISTS effective_policies_pkey;
CREATE UNIQUE INDEX IF NOT EXISTS effective_policies_pkey ON effective_policies (policy_id);

ALTER TABLE effective_policies DROP COLUMN IF EXISTS deleted_at;

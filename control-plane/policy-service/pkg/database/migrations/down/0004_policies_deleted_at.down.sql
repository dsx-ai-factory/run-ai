-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- since deleted_at is currently a part of an index, we must remove the rows that have it set before changing the index.
-- otherwise, we'll fail due to multiple rows with the same (tenant_id, workload_type, name) tuple, which is supposed
-- to be unique.
DO $$
BEGIN
    IF EXISTS (
        SELECT column_name
        FROM information_schema.columns
        WHERE table_name = 'policies'
        AND column_name = 'deleted_at'
    ) THEN
        DELETE FROM policies WHERE deleted_at IS NOT NULL;
    END IF;
END $$;

DROP INDEX IF EXISTS by_name;
CREATE UNIQUE INDEX IF NOT EXISTS by_name ON policies (tenant_id, workload_type, name);

ALTER TABLE policies DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE policies DROP COLUMN IF EXISTS deleted_by;



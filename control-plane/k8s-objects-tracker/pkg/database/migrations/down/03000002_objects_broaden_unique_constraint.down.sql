-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Reverse the per-type split: drop the two partial unique indexes and restore
-- the single constraint that was in effect before 03000002.
--
-- The constraint add is wrapped in a conditional block because Postgres has no
-- native ADD CONSTRAINT IF NOT EXISTS; this makes the down migration safe to
-- re-run without failing on an already-restored constraint, matching the
-- DO-block convention used in 02201002 and 02171001.

DROP INDEX IF EXISTS unique_resource_claim_templates;
DROP INDEX IF EXISTS unique_cluster_scoped_objects;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'unique_tenant_type_cluster_name_project'
          AND conrelid = 'objects'::regclass
    ) THEN
        ALTER TABLE objects
            ADD CONSTRAINT unique_tenant_type_cluster_name_project
            UNIQUE (tenant_id, type, cluster_id, name, project_id);
    END IF;
END $$;

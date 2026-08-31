-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


-- remove empty records. they are auto-created as part of service runs, and it will reduce
-- the number of records to migrate.
DELETE FROM policies WHERE rules_v2 is NULL AND defaults_v2 IS NULL and imposed_assets IS NULL;

-- delete duplicate policies
DELETE FROM policies
WHERE policies.ctid NOT IN (
    SELECT min(p.ctid)
    FROM policies p
    GROUP BY (
              p.workload_type,
              p.scope,
              p.tenant_id,
              p.department_id,
              p.project_id
            )
);

-- sanitize the effective policies table to sync with the policies table
DELETE FROM effective_policies WHERE policy_id NOT IN (select id FROM policies);

-- create unique index for type + scope
CREATE UNIQUE INDEX IF NOT EXISTS policies_by_type_and_scope 
	ON policies(
        workload_type,
        scope,
        COALESCE(tenant_id, 0),
        COALESCE(cluster_id::text, ''),
        COALESCE(department_id::text, ''),
        COALESCE(project_id, 0));


-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Migrate roles without permission sets to use permission sets
-- For each role that doesn't have a permission_set in role_to_permission_set:
-- 1. Create a new permission_set with internal=true and name based on role name and tenant_id
-- 2. Copy the role's permissions to role_permissions linked to the new permission_set
-- 3. Link the role to the new permission_set in role_to_permission_set

-- First, identify roles that need migration
WITH roles_to_migrate AS (
    SELECT
        r.id AS role_id,
        r.name AS role_name,
        r.tenant_id,
        CASE
            WHEN r.tenant_id IS NOT NULL THEN CONCAT(r.tenant_id, '-', r.name, '-internal-ps')
            ELSE CONCAT(r.name, '-internal-ps')
            END AS permission_set_name
    FROM roles r
    WHERE NOT EXISTS (
        SELECT 1
        FROM role_to_permission_set rps
        WHERE rps.role_id = r.id
    )
),
-- Create permission sets for roles that don't have one
     created_permission_sets AS (
         INSERT INTO permission_sets (id, name, description, internal, created_at, updated_at)
             SELECT
                 gen_random_uuid() AS id,
                 permission_set_name AS name,
                 '' AS description,
                 true AS internal,
                 NOW() AS created_at,
                 NOW() AS updated_at
             FROM roles_to_migrate
             ON CONFLICT (name) DO NOTHING
             RETURNING id, name
     ),
-- Copy permissions from old permissions table to role_permissions for the new permission sets
     copied_permissions AS (
         INSERT INTO role_permissions (id, created_at, updated_at, permission_set_id, resource_type, "create", read, update, delete, sync)
             SELECT
                 gen_random_uuid() AS id,
                 NOW() AS created_at,
                 NOW() AS updated_at,
                 cps.id AS permission_set_id,
                 p.resource_type,
                 p."create",
                 p.read,
                 p.update,
                 p.delete,
                 p.sync
             FROM permissions p
                      INNER JOIN roles_to_migrate rtm ON p.role_id = rtm.role_id
                      INNER JOIN created_permission_sets cps ON cps.name = rtm.permission_set_name
             ON CONFLICT (permission_set_id, resource_type) DO NOTHING
             RETURNING permission_set_id
     )
-- Link roles to their new permission sets
INSERT INTO role_to_permission_set (role_id, permission_set_id)
SELECT
    rtm.role_id,
    cps.id AS permission_set_id
FROM roles_to_migrate rtm
         INNER JOIN created_permission_sets cps ON cps.name = rtm.permission_set_name
ON CONFLICT (role_id, permission_set_id) DO NOTHING;
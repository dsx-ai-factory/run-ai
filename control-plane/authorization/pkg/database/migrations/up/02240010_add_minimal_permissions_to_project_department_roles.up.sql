-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Migration to add clusters-minimal and nodepools-minimal permissions to custom roles and permission sets
-- that have department or project read permissions but are missing the minimal permissions

-- Add clusters-minimal read permission to custom roles that have department OR project read permission
-- but don't yet have clusters-minimal permission
INSERT INTO permissions (role_id, resource_type, "create", read, update, delete, sync, created_at, updated_at)
SELECT DISTINCT
    p.role_id,
    'clusters-minimal' AS resource_type,
    false AS "create",
    true AS read,
    false AS update,
    false AS delete,
    false AS sync,
    NOW() AS created_at,
    NOW() AS updated_at
FROM permissions p
INNER JOIN roles r ON r.id = p.role_id
WHERE p.resource_type IN ('department', 'project')
  AND p.read = true
  AND r.custom = true
  AND NOT EXISTS (
    SELECT 1 
    FROM permissions p2 
    WHERE p2.role_id = p.role_id 
      AND p2.resource_type = 'clusters-minimal'
  )
ON CONFLICT (role_id, resource_type) DO NOTHING;

-- Add nodepools-minimal read permission to custom roles that have department OR project read permission
-- but don't yet have nodepools-minimal permission
INSERT INTO permissions (role_id, resource_type, "create", read, update, delete, sync, created_at, updated_at)
SELECT DISTINCT
    p.role_id,
    'nodepools-minimal' AS resource_type,
    false AS "create",
    true AS read,
    false AS update,
    false AS delete,
    false AS sync,
    NOW() AS created_at,
    NOW() AS updated_at
FROM permissions p
INNER JOIN roles r ON r.id = p.role_id
WHERE p.resource_type IN ('department', 'project')
  AND p.read = true
  AND r.custom = true
  AND NOT EXISTS (
    SELECT 1 
    FROM permissions p2 
    WHERE p2.role_id = p.role_id 
      AND p2.resource_type = 'nodepools-minimal'
  )
ON CONFLICT (role_id, resource_type) DO NOTHING;

-- Add clusters-minimal read permission to permission sets that have department OR project read permission
-- but don't yet have clusters-minimal permission (excluding internal/static permission sets)
INSERT INTO role_permissions (id, permission_set_id, resource_type, "create", read, update, delete, sync, created_at, updated_at)
SELECT DISTINCT
    gen_random_uuid() AS id,
    rp.permission_set_id,
    'clusters-minimal' AS resource_type,
    false AS "create",
    true AS read,
    false AS update,
    false AS delete,
    false AS sync,
    NOW() AS created_at,
    NOW() AS updated_at
FROM role_permissions rp
INNER JOIN permission_sets ps ON ps.id = rp.permission_set_id
WHERE rp.resource_type IN ('department', 'project')
  AND rp.read = true
  AND NOT EXISTS (
    SELECT 1 
    FROM role_permissions rp2 
    WHERE rp2.permission_set_id = rp.permission_set_id 
      AND rp2.resource_type = 'clusters-minimal'
  )
ON CONFLICT (permission_set_id, resource_type) DO NOTHING;

-- Add nodepools-minimal read permission to permission sets that have department OR project read permission
-- but don't yet have nodepools-minimal permission (excluding internal/static permission sets)
INSERT INTO role_permissions (id, permission_set_id, resource_type, "create", read, update, delete, sync, created_at, updated_at)
SELECT DISTINCT
    gen_random_uuid() AS id,
    rp.permission_set_id,
    'nodepools-minimal' AS resource_type,
    false AS "create",
    true AS read,
    false AS update,
    false AS delete,
    false AS sync,
    NOW() AS created_at,
    NOW() AS updated_at
FROM role_permissions rp
INNER JOIN permission_sets ps ON ps.id = rp.permission_set_id
WHERE rp.resource_type IN ('department', 'project')
  AND rp.read = true
  AND NOT EXISTS (
    SELECT 1 
    FROM role_permissions rp2 
    WHERE rp2.permission_set_id = rp.permission_set_id 
      AND rp2.resource_type = 'nodepools-minimal'
  )
ON CONFLICT (permission_set_id, resource_type) DO NOTHING;

-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Replace the deprecated 'apps' resource type with 'service-account'.
--
-- Two tables store a resource type next to a set of flags (create/read/update/delete/sync):
--   permissions       one row per (role_id, resource_type)
--   role_permissions  one row per (permission_set_id, resource_type)   -- a permission set's permissions
-- Both have a unique key on (owner, resource_type), so a plain rename of 'apps' to 'service-account' fails for
-- any owner that already has BOTH rows. That is why each table is migrated in two steps:
--
--   step 1  delete the 'apps' row of every owner that already has a 'service-account' row
--           (the existing 'service-account' row is kept exactly as it is)
--   step 2  rename the remaining 'apps' rows to 'service-account'
--           (after step 1 these owners have no 'service-account' row, so the rename cannot collide)
--
-- Example, role A has both rows, role B only has 'apps':
--   before   A: apps(read)  A: service-account(read)  B: apps(read)
--   step 1   (A: apps deleted)
--   step 2   B: apps -> service-account
--   after    A: service-account(read)  B: service-account(read)
--
-- Running it by hand: the tables live in the "authorization" schema (quoted, AUTHORIZATION is a reserved word) and
-- the migrator sets that schema for you, a plain SQL client does not. Run it in one transaction:
--   BEGIN;
--   SET LOCAL search_path TO "authorization";
--   -- preview: rows that step 1 will delete / step 2 will rename
--   SELECT count(*) FROM permissions WHERE resource_type = 'apps';
--   -- (now run the statements below)
--   COMMIT;   -- or ROLLBACK; to undo

-- permissions (owner: role_id) ---------------------------------------------------------------------------------

-- step 1: drop 'apps' where the same role already has 'service-account'
DELETE FROM permissions
WHERE resource_type = 'apps'
  AND EXISTS (
      SELECT 1
      FROM permissions AS existing
      WHERE existing.role_id = permissions.role_id
        AND existing.resource_type = 'service-account'
  );

-- step 2: rename what is left
UPDATE permissions
SET resource_type = 'service-account'
WHERE resource_type = 'apps';

-- role_permissions (owner: permission_set_id) ------------------------------------------------------------------

-- step 1: drop 'apps' where the same permission set already has 'service-account'
DELETE FROM role_permissions
WHERE resource_type = 'apps'
  AND EXISTS (
      SELECT 1
      FROM role_permissions AS existing
      WHERE existing.permission_set_id = role_permissions.permission_set_id
        AND existing.resource_type = 'service-account'
  );

-- step 2: rename what is left
UPDATE role_permissions
SET resource_type = 'service-account'
WHERE resource_type = 'apps';

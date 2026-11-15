-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Replace the deprecated 'apps' resource type with 'service-account' in role_permissions (the permissions of a
-- permission set; this is the table access checks read through role_to_permission_set).
-- The legacy per-role "permissions" table is not read anymore and is going to be dropped, so it is not migrated.
--
-- role_permissions has one row per (permission_set_id, resource_type), with the flags create/read/update/delete/sync.
-- That pair is a unique key, so a plain rename of 'apps' to 'service-account' fails for any permission set that
-- already has BOTH rows. That is why the migration is done in two steps:
--
--   step 1  delete the 'apps' row of every permission set that already has a 'service-account' row
--           (the existing 'service-account' row is kept exactly as it is)
--   step 2  rename the remaining 'apps' rows to 'service-account'
--           (after step 1 these permission sets have no 'service-account' row, so the rename cannot collide)
--
-- Example, permission set A has both rows, permission set B only has 'apps':
--   before   A: apps(read)  A: service-account(read)  B: apps(read)
--   step 1   (A: apps deleted)
--   step 2   B: apps -> service-account
--   after    A: service-account(read)  B: service-account(read)
--
-- Running it by hand: the table lives in the "authorization" schema (quoted, AUTHORIZATION is a reserved word) and
-- the migrator sets that schema for you, a plain SQL client does not. Run it in one transaction:
--   BEGIN;
--   SET LOCAL search_path TO "authorization";
--   -- preview: how many rows step 1 / step 2 will touch
--   SELECT count(*) FROM role_permissions WHERE resource_type = 'apps';
--   -- (now run the statements below)
--   COMMIT;   -- or ROLLBACK; to undo

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

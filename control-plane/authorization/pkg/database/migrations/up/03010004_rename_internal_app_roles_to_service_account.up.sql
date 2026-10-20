-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Rename the internal 'Cluster application' and 'Installer application' roles, and their permission sets, to
-- '... service account'. Renaming in place keeps the ids, so access rules, role-to-permission-set links and rules
-- already synced to clusters keep pointing at the same rows. The static initializers upsert by name, so without the
-- rename they would create new rows and leave the old ones orphaned. The NOT EXISTS guards skip a rename when a row
-- with the new name already exists, which would otherwise violate the unique name.
UPDATE permission_sets
SET name = 'Cluster service account permission set'
WHERE name = 'Cluster app permission set'
  AND NOT EXISTS (SELECT 1 FROM permission_sets WHERE name = 'Cluster service account permission set');

UPDATE permission_sets
SET name = 'Installer service account permission set'
WHERE name = 'Installer app permission set'
  AND NOT EXISTS (SELECT 1 FROM permission_sets WHERE name = 'Installer service account permission set');

UPDATE roles
SET name = 'Cluster service account'
WHERE name = 'Cluster application'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Cluster service account');

UPDATE roles
SET name = 'Installer service account'
WHERE name = 'Installer application'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Installer service account');

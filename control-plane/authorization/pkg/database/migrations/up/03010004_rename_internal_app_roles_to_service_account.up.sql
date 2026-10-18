-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Rename the internal 'Cluster application' and 'Installer application' roles to '... service account'.
-- Renaming in place keeps the role ids, so access rules, permission-set links and rules already synced to clusters
-- keep pointing at the same role. The NOT EXISTS guards skip a rename when a role with the new name was already
-- created (e.g. by the static role initializer), which would otherwise violate the unique role name.
UPDATE roles
SET name = 'Cluster service account'
WHERE name = 'Cluster application'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Cluster service account');

UPDATE roles
SET name = 'Installer service account'
WHERE name = 'Installer application'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Installer service account');

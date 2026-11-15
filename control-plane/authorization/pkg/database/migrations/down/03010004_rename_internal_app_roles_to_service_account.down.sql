-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE permission_sets
SET name = 'Cluster app permission set'
WHERE name = 'Cluster service account permission set'
  AND NOT EXISTS (SELECT 1 FROM permission_sets WHERE name = 'Cluster app permission set');

UPDATE permission_sets
SET name = 'Installer app permission set'
WHERE name = 'Installer service account permission set'
  AND NOT EXISTS (SELECT 1 FROM permission_sets WHERE name = 'Installer app permission set');

UPDATE roles
SET name = 'Cluster application'
WHERE name = 'Cluster service account'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Cluster application');

UPDATE roles
SET name = 'Installer application'
WHERE name = 'Installer service account'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Installer application');

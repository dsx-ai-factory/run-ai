-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

UPDATE roles
SET name = 'Cluster application'
WHERE name = 'Cluster service account'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Cluster application');

UPDATE roles
SET name = 'Installer application'
WHERE name = 'Installer service account'
  AND NOT EXISTS (SELECT 1 FROM roles WHERE name = 'Installer application');

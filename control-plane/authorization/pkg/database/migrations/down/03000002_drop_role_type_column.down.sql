-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Restores the column and its values. Only ever three roles carried 'NON_DGXC' and five
-- carried 'DGXC'; everything else - static and custom alike - took the 'ALL' default,
-- which the column definition reinstates.
--
-- The five DGXC roles are not restored here: 03000001 renamed them, so they do not
-- carry these names yet. Its own down migration renames them back and sets their
-- role_type at the same time.

ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS role_type varchar default 'ALL';

UPDATE roles
SET role_type = 'NON_DGXC'
WHERE scope_type = 'system'
  AND name IN ('Cloud operator',
               'Department administrator',
               'Department administrator (Legacy)');

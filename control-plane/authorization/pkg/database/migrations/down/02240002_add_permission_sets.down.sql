-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS roles DROP COLUMN IF EXISTS no_permission_set;

DROP TABLE IF EXISTS role_to_permission_set;

ALTER TABLE IF EXISTS permissions DROP CONSTRAINT IF EXISTS fk_permission_set_permissions;
ALTER TABLE IF EXISTS permissions DROP COLUMN IF EXISTS permission_set_id;

DROP TABLE IF EXISTS permission_sets;



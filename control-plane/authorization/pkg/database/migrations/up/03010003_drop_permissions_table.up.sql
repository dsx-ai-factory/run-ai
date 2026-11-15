-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The legacy per-role permissions table. Since 02240007 every role carries its permissions
-- in permission sets, and evaluation, the permissions summary and role reads all go through
-- role_to_permission_set -> role_permissions. Nothing has read this table since, so dropping
-- it changes no role's effective permissions.
--
-- Ship this only in a release whose code no longer maps models.Role.Permissions to this
-- table. Pods still running older code fail static role initialisation on restart, and
-- fail updates to custom roles created with a permission list, with
-- 'relation "permissions" does not exist' once this applies.

DROP TABLE IF EXISTS permissions;

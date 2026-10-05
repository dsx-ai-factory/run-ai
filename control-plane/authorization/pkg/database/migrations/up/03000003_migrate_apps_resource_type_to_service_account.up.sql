-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The deprecated 'apps' resource type is replaced by 'service-account'.
-- Both tables are keyed by (owner, resource_type), so where an owner already has a 'service-account' row the
-- redundant 'apps' row is dropped (that row stays as it is) and otherwise 'apps' is renamed.

-- permissions: owner is role_id
DELETE FROM permissions apps
    USING permissions service_account
WHERE apps.resource_type = 'apps'
  AND service_account.resource_type = 'service-account'
  AND apps.role_id = service_account.role_id;

UPDATE permissions
SET resource_type = 'service-account'
WHERE resource_type = 'apps';

-- role_permissions (permission sets): owner is permission_set_id
DELETE FROM role_permissions apps
    USING role_permissions service_account
WHERE apps.resource_type = 'apps'
  AND service_account.resource_type = 'service-account'
  AND apps.permission_set_id = service_account.permission_set_id;

UPDATE role_permissions
SET resource_type = 'service-account'
WHERE resource_type = 'apps';

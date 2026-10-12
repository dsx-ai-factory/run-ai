-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Recreates the table empty, as 02150001 defined it, so code that still maps
-- models.Role.Permissions to it can start. The rows are not restored: nothing read them,
-- and every role's permissions live in role_permissions.

CREATE TABLE IF NOT EXISTS permissions (
    created_at    timestamp with time zone,
    updated_at    timestamp with time zone,
    role_id       bigint not null
        constraint fk_roles_permissions
            references roles
            on delete cascade,
    resource_type text   not null,
    "create"      boolean,
    read          boolean,
    update        boolean,
    delete        boolean,
    sync          boolean,
    primary key (role_id, resource_type)
);

CREATE INDEX IF NOT EXISTS idx_permissions_resource_type
    ON permissions (resource_type);

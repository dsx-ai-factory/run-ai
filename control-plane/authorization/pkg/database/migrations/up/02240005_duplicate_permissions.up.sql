-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS permission_sets ADD COLUMN IF NOT EXISTS created_by VARCHAR;
ALTER TABLE IF EXISTS permission_sets ADD COLUMN IF NOT EXISTS updated_by VARCHAR;
ALTER TABLE IF EXISTS permission_sets ADD CONSTRAINT permission_sets_name_key UNIQUE (name);

CREATE TABLE IF NOT EXISTS role_permissions (
    id                  UUID PRIMARY KEY,
    created_at          TIMESTAMPTZ,
    updated_at          TIMESTAMPTZ,
    permission_set_id   UUID NOT NULL
       CONSTRAINT fk_permission_set_permissions
           REFERENCES permission_sets
           ON DELETE CASCADE,
    resource_type       VARCHAR   NOT NULL,
    "create"            BOOLEAN,
    read                BOOLEAN,
    update              BOOLEAN,
    delete              BOOLEAN,
    sync                BOOLEAN,
    CONSTRAINT unique_permission_set_resource_type UNIQUE (permission_set_id, resource_type)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_resource_type
    ON role_permissions (resource_type);

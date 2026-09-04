-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS permission_sets (
    id              UUID PRIMARY KEY,
    name            VARCHAR NOT NULL,
    description     VARCHAR,
    internal        BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ
);

ALTER TABLE IF EXISTS permissions ADD COLUMN IF NOT EXISTS permission_set_id UUID;
ALTER TABLE IF EXISTS permissions ADD CONSTRAINT fk_permission_set_permissions FOREIGN KEY(permission_set_id) REFERENCES permission_sets(id);

CREATE TABLE IF NOT EXISTS role_to_permission_set (
    role_id             bigserial NOT NULL,
    permission_set_id   UUID NOT NULL,
    CONSTRAINT fk_role FOREIGN KEY(role_id) REFERENCES roles(id),
    CONSTRAINT fk_permission_set FOREIGN KEY(permission_set_id) REFERENCES permission_sets(id)
);

ALTER TABLE IF EXISTS roles ADD COLUMN no_permission_set BOOLEAN NOT NULL DEFAULT true;
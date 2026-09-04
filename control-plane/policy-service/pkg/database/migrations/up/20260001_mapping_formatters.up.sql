-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS mapping_formatters (
    id          UUID        NOT NULL PRIMARY KEY,

    name        text        NOT NULL,
    description text,

    scope_type  text        NOT NULL CHECK (scope_type IN ('system', 'tenant')),
    tenant_id   integer     NOT NULL DEFAULT 0,

    fields      jsonb       NOT NULL,

    created_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_by  text,
    updated_at  timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT mapping_formatters_scope_tenant_id
    CHECK ((scope_type = 'system' AND tenant_id = 0) OR
           (scope_type = 'tenant' AND tenant_id <> 0)),

    CONSTRAINT mapping_formatters_name_prefix_by_scope
    CHECK ((scope_type = 'tenant' AND name !~ '^runai-') OR
           (scope_type = 'system' AND name ~ '^runai-'))
    );

CREATE UNIQUE INDEX IF NOT EXISTS mapping_formatters_unique
    ON mapping_formatters (name, scope_type, tenant_id);

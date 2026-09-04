-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE assets_metadata ADD COLUMN IF NOT EXISTS private BOOLEAN DEFAULT FALSE;

DROP INDEX IF EXISTS assets_metadata_by_kind_name_tenant;

CREATE UNIQUE INDEX assets_metadata_by_kind_name_tenant
    ON assets_metadata (kind, tenant_id, name)
    WHERE scope = 'tenant'::text AND COALESCE(private, FALSE) = FALSE;

CREATE UNIQUE INDEX assets_metadata_by_kind_name_tenant_private
    ON assets_metadata (kind, tenant_id, name, created_by)
    WHERE scope = 'tenant'::text AND COALESCE(private, FALSE) = TRUE;
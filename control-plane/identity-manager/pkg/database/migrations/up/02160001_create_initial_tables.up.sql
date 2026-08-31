-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS realms (
    id                  UUID NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    created_by          TEXT,
    updated_by          TEXT,
    deleted_by          TEXT,
    realm               TEXT NOT NULL ,
    tenant_id           TEXT NOT NULL PRIMARY KEY
);
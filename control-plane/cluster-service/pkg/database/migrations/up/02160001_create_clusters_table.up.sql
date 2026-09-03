-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS cluster (
    id            UUID PRIMARY KEY,
    name          TEXT,
    tenant_id     INTEGER,
    domain        TEXT,
    version       TEXT,
    created_at    TIMESTAMPTZ DEFAULT now(),
    updated_at    TIMESTAMPTZ DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    last_liveness TIMESTAMPTZ,
    status        JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_tenant ON cluster (tenant_id);
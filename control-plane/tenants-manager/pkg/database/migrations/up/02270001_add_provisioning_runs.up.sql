-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS provisioning_runs (
    id           SERIAL PRIMARY KEY,
    tenant_id    INTEGER      NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    started_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    ended_at     TIMESTAMPTZ,
    status       VARCHAR(50)  NOT NULL DEFAULT 'running',
    diagnostics  TEXT
);

CREATE INDEX IF NOT EXISTS idx_provisioning_runs_tenant_id ON provisioning_runs(tenant_id);

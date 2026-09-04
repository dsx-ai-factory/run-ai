-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS priorities (
    id            UUID primary key,
    name          VARCHAR NOT NULL,
    value         INTEGER NOT NULL,
    source        TEXT,
    scope_type    TEXT NOT NULL,
    scope_id      TEXT,
    tenant_id     TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    created_by    TEXT,
    updated_by    TEXT,
    deleted_by    TEXT
);
CREATE UNIQUE INDEX priorities_tenant_unique ON priorities (name, scope_type, scope_id) WHERE scope_id IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX priorities_system_unique ON priorities (name, scope_type) WHERE scope_id IS NULL AND deleted_at IS NULL;

INSERT INTO priorities ("id","name", "value", "source", "scope_type", "created_by", "updated_by") VALUES
    (gen_random_uuid(), 'very-low', 25, 'Run:ai', 'system', 'Run:ai', 'Run:ai'),
    (gen_random_uuid(), 'low', 40, 'Run:ai', 'system', 'Run:ai', 'Run:ai'),
    (gen_random_uuid(), 'medium-low', 65, 'Run:ai', 'system', 'Run:ai', 'Run:ai'),
    (gen_random_uuid(), 'medium', 80, 'Run:ai', 'system', 'Run:ai', 'Run:ai'),
    (gen_random_uuid(), 'medium-high', 90, 'Run:ai', 'system', 'Run:ai', 'Run:ai'),
    (gen_random_uuid(), 'high', 125, 'Run:ai', 'system', 'Run:ai', 'Run:ai'),
    (gen_random_uuid(), 'very-high', 150, 'Run:ai', 'system', 'Run:ai', 'Run:ai');

CREATE TABLE IF NOT EXISTS priority_statuses (
    id              UUID primary key,
    priority_id     UUID NOT NULL,
    cluster_id      TEXT NOT NULL,
    phase           TEXT,
    phase_message   TEXT,

    CONSTRAINT fk_priority FOREIGN KEY(priority_id) REFERENCES priorities(id)
)

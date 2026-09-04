-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS settings (
    id                          UUID NOT NULL PRIMARY KEY,
    tenant_id                   TEXT NOT NULL,
    key                         TEXT NOT NULL,
    value                       JSONB,
    updated_at                  TIMESTAMPTZ,
    updated_by_customer         BOOLEAN NOT NULL
);
ALTER TABLE settings DROP CONSTRAINT if exists unique_setting;
ALTER TABLE settings ADD CONSTRAINT unique_setting UNIQUE (tenant_id, key);

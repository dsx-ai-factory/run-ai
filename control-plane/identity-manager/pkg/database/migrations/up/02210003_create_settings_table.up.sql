-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS realm_settings (
    id                          UUID NOT NULL PRIMARY KEY,
    tenant_id                   TEXT NOT NULL,
    setting_key                 TEXT NOT NULL,
    setting_value               JSONB,
    updated_at                  TIMESTAMPTZ,
    updated_by_customer         BOOLEAN NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES realms(tenant_id)
);

ALTER TABLE realm_settings ADD CONSTRAINT fk_tenant_id  FOREIGN KEY (tenant_id) REFERENCES realms(tenant_id) ON DELETE CASCADE;